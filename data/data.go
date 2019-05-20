package data

import (
	"bytes"
	"crypto"
	_ "crypto/sha512"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/boltdb/bolt"
	scrypto "github.com/boreq/starlight/crypto"
	"github.com/boreq/starlight/network/node"
	"github.com/pkg/errors"
)

// SigningHash specifies the hash used for generating the signature.
const SigningHash = crypto.SHA512

// minNickLength specifies the min length of a nick.
const minNickLength = 3

// maxNickLength specifies the max length of a nick.
const maxNickLength = 20

// nickRegexp is used to validate nicks.
var nickRegexp = regexp.MustCompile(`^[a-zA-Z]{1}[a-zA-Z0-9\_\-\[\]]+$`)

// NickData represents an intent to set a nickname.
type NickData struct {
	Id        node.ID   `json:"id"`
	Nick      string    `json:"nick"`
	Time      time.Time `json:"time"`
	PublicKey []byte    `json:"publicKey"`
	Signature []byte    `json:"signature"`
}

// GetDataToSign returns the data which should be signed to produce the
// signature.
func (n NickData) GetDataToSign() []byte {
	buf := &bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("%d", n.Time.Unix()))
	buf.Write(n.Id)
	buf.WriteString(n.Nick)
	return buf.Bytes()
}

// Validate checks if this struct is filled correctly.
func (n NickData) Validate() error {
	// Public key
	publicKey, err := scrypto.NewPublicKey(n.PublicKey)
	if err != nil {
		return errors.Wrap(err, "could not read the public key")
	}

	// Id
	if !node.ValidateId(n.Id) {
		return errors.New("id is invalid")
	}
	id, err := publicKey.Hash()
	if err != nil {
		return errors.Wrap(err, "could not hash the public key")
	}
	if !node.CompareId(id, n.Id) {
		return errors.New("id does not match the public key")
	}

	// Nick
	if err := ValidateNick(n.Nick); err != nil {
		return errors.Wrap(err, "invalid nick")
	}

	// Time
	if isZero := n.Time.IsZero(); isZero {
		return errors.New("time is zero")
	}

	// Signature
	data := n.GetDataToSign()
	if err := publicKey.Validate(data, n.Signature, SigningHash); err != nil {
		return errors.Wrap(err, "could not validate the signature")
	}

	return nil
}

// ValidateNick checks if the nick is valid.
func ValidateNick(nick string) error {
	if len(nick) < minNickLength {
		return errors.Errorf("nick needs to be at least %d characters long", minNickLength)
	}
	if len(nick) > maxNickLength {
		return errors.Errorf("nick needs to be at most %d characters long", maxNickLength)
	}
	if result := nickRegexp.MatchString(nick); !result {
		return errors.Errorf("nick does not match the regular expression")
	}
	return nil
}

var InvalidNickDataErr = errors.New("invalid nick data")
var NewerNickDataPresentErr = errors.New("newer nick data is available")
var NickConflictErr = errors.New("nick is already taken")
var InvalidNodeIdErr = errors.New("invalid node id")

const nickDataBucket = "nickdata"
const nicksBucket = "nicks"

// NewBoltRepository opens or creates a repository using bolt as an underlying
// storage.
func NewBoltRepository(path string) (*BoltRepository, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not open the database")
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(nickDataBucket)); err != nil {
			return errors.Wrap(err, "nickDataBucket creation failed")
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(nicksBucket)); err != nil {
			return errors.Wrap(err, "nicksBucket creation failed")
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "could not create the bucket")
	}

	rv := &BoltRepository{
		db: db,
	}
	return rv, nil
}

type BoltRepository struct {
	db *bolt.DB
}

// List returns a list of all stored entires.
func (r *BoltRepository) List() ([]NickData, error) {
	rv := make([]NickData, 0)
	if err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(nickDataBucket))
		return b.ForEach(func(k, v []byte) error {
			nickData, err := r.unmarshalNickData(v)
			if err != nil {
				return errors.Wrap(err, "unmarshal failed")
			}
			rv = append(rv, *nickData)
			return nil
		})
	}); err != nil {
		return nil, err
	}
	return rv, nil
}

// Get returns an entry for a specific node id. If the node id is invalid
// InvalidNodeIdErr is returned. If the entry doesn't exist nil is returned
// without an error.
func (r *BoltRepository) Get(id node.ID) (*NickData, error) {
	if !node.ValidateId(id) {
		return nil, InvalidNodeIdErr
	}

	var nickData *NickData = nil
	if err := r.db.View(func(tx *bolt.Tx) error {
		nd, err := r.getNickData(tx, id)
		if err != nil {
			return err
		}
		nickData = nd
		return nil
	}); err != nil {
		return nil, err
	}
	return nickData, nil
}

func (r *BoltRepository) getNickData(tx *bolt.Tx, id node.ID) (*NickData, error) {
	b := tx.Bucket([]byte(nickDataBucket))
	v := b.Get(id)
	if v != nil {
		return r.unmarshalNickData(v)
	}
	return nil, nil
}

func (r *BoltRepository) unmarshalNickData(data []byte) (*NickData, error) {
	nickData := &NickData{}
	if err := json.Unmarshal(data, nickData); err != nil {
		return nil, errors.Wrap(err, "json unmarshal failed")
	}
	return nickData, nil
}

// Put inserts a new entry. In case of a nick collision with a different node
// NickConflictErr is returned. In case the entry is invalid InvalidNickDataErr
// is returned. In case there is a newer nick data available for this node
// NewerNickDataPresentErr is returned.
func (r *BoltRepository) Put(nickData *NickData) error {
	if err := nickData.Validate(); err != nil {
		return InvalidNickDataErr
	}

	value, err := json.Marshal(nickData)
	if err != nil {
		return errors.Wrap(err, "marshaling nick data failed")
	}

	if err := r.db.Update(func(tx *bolt.Tx) error {
		// Confirm that the nick doesn't exist
		nicksB := tx.Bucket([]byte(nicksBucket))
		existingId := nicksB.Get([]byte(nickData.Nick))
		if existingId != nil {
			if !node.CompareId(existingId, nickData.Id) {
				return NickConflictErr
			}
		}

		// Confirm that there is no newer nick data
		previousNickData, err := r.getNickData(tx, nickData.Id)
		if err != nil {
			return errors.Wrap(err, "error retrieving the previous nick data")
		}
		if previousNickData != nil {
			if previousNickData.Time.After(nickData.Time) {
				return NewerNickDataPresentErr
			}
		}

		// Insert new nick
		if err := nicksB.Put(nickData.Id, value); err != nil {
			return errors.Wrap(err, "nicks bucket put failed")
		}

		nickDataB := tx.Bucket([]byte(nickDataBucket))
		if err := nickDataB.Put(nickData.Id, value); err != nil {
			return errors.Wrap(err, "nick data bucket put failed")
		}
		return nil
	}); err != nil {
		if err == NickConflictErr || err == NewerNickDataPresentErr {
			return err
		}
		return errors.Wrap(err, "update failed")
	}
	return nil
}

// Close closes the database.
func (r *BoltRepository) Close() error {
	return r.db.Close()
}
