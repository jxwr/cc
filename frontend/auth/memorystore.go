package auth

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"time"
)

type MemoryTokenStore struct {
	tokens   map[string]*MemoryToken
	idTokens map[string]*MemoryToken
	salt     string
}

type MemoryToken struct {
	ExpireAt time.Time
	Token    string
	Id       string
}

func (t *MemoryToken) IsExpired() bool {
	return time.Now().After(t.ExpireAt)
}

func (t *MemoryToken) String() string {
	return t.Token
}

/* lookup 'exp' or 'id' */
func (t *MemoryToken) Claims(key string) interface{} {
	switch key {
	case "exp":
		return t.ExpireAt
	case "id":
		return t.Id
	case "token":
		return t.Token
	default:
		return nil
	}
}

func GenerateToken(id string) string {
	hash := sha1.New()
	now := time.Now()
	timeStr := now.Format(time.ANSIC)
	hash.Write([]byte(timeStr))
	hash.Write([]byte(id))
	hash.Write([]byte("salt"))
	return base64.URLEncoding.EncodeToString(hash.Sum(nil))
}

func (s *MemoryTokenStore) generateToken(id string) []byte {
	hash := sha1.New()
	now := time.Now()
	timeStr := now.Format(time.ANSIC)
	hash.Write([]byte(timeStr))
	hash.Write([]byte(id))
	hash.Write([]byte("salt"))
	return hash.Sum(nil)
}

/* returns a new token with specific id */
func (s *MemoryTokenStore) NewToken(id interface{}) *MemoryToken {
	strId := id.(string)
	bToken := s.generateToken(strId)
	strToken := base64.URLEncoding.EncodeToString(bToken)
	t := &MemoryToken{
		ExpireAt: time.Now().Add(time.Hour * 12),
		Token:    strToken,
		Id:       strId,
	}
	oldT, ok := s.idTokens[strId]
	if ok {
		delete(s.tokens, oldT.Token)
	}
	s.tokens[strToken] = t
	s.idTokens[strId] = t
	return t
}

func (s *MemoryTokenStore) UpdateToken(id, token string) *MemoryToken {
	t := &MemoryToken{
		ExpireAt: time.Now().Add(time.Hour * 12),
		Token:    token,
		Id:       id,
	}
	oldT, ok := s.idTokens[id]
	if ok {
		delete(s.tokens, oldT.Token)
	}
	s.tokens[token] = t
	s.idTokens[id] = t
	return t
}

/* Create a new memory store */
func NewTokenStore(salt string) *MemoryTokenStore {
	return &MemoryTokenStore{
		salt:     salt,
		tokens:   make(map[string]*MemoryToken),
		idTokens: make(map[string]*MemoryToken),
	}

}

func (s *MemoryTokenStore) DeleteIdToken(id string) {
	_, ok := s.idTokens[id]
	if !ok {
		return
	}
	delete(s.idTokens, id)
}

func (s *MemoryTokenStore) CheckIdToken(id, strToken string) (*MemoryToken, bool, error) {
	t, ok := s.idTokens[id]
	if !ok {
		return nil, false, errors.New("Token not exist")
	}
	if t.String() != strToken {
		return nil, true, errors.New("Failed to authenticate")
	}

	if t.ExpireAt.Before(time.Now()) {
		delete(s.idTokens, id)
		return nil, true, errors.New("Token expired")
	}
	return t, true, nil
}

func (s *MemoryTokenStore) CheckToken(strToken string) (*MemoryToken, error) {
	t, ok := s.tokens[strToken]
	if !ok {
		return nil, errors.New("Failed to authenticate")
	}
	if t.ExpireAt.Before(time.Now()) {
		delete(s.tokens, strToken)
		return nil, errors.New("Token expired")
	}
	return t, nil
}
