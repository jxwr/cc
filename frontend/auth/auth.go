package auth

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ksarch-saas/cc/frontend/api"
	"github.com/ksarch-saas/cc/meta"
	"net/http"
)

type TokenAuth struct {
	handler http.Handler
	store   *MemoryTokenStore
	getter  TokenGetter
}

type TokenGetter interface {
	GetUserFromRequest(req *http.Request) string
	GetTokenFromRequest(req *http.Request) string
}

type ClaimSetter interface {
	SetClaim(string, interface{}) ClaimSetter
}

type ClaimGetter interface {
	Claims(string) interface{}
}

type QueryStringTokenGetter struct {
	User  string
	Token string
}

func (q QueryStringTokenGetter) GetTokenFromRequest(req *http.Request) string {
	authStr := req.Header.Get(q.Token)
	return authStr
}

func (q QueryStringTokenGetter) GetUserFromRequest(req *http.Request) string {
	authStr := req.Header.Get(q.User)
	return authStr
}

func NewQueryStringTokenGetter(user, token string) *QueryStringTokenGetter {
	return &QueryStringTokenGetter{
		User:  user,
		Token: token,
	}
}

/*
	Returns a TokenAuth object implemting Handler interface

	if a handler is given it proxies the request to the handler

	store is the TokenStore that stores and verify the tokens
*/
func NewTokenAuth(handler http.Handler, store *MemoryTokenStore, getter TokenGetter) *TokenAuth {
	t := &TokenAuth{
		handler: handler,
		store:   store,
		getter:  getter,
	}
	if t.getter == nil {
		t.getter = NewQueryStringTokenGetter("User", "Token")
	}
	return t
}

/* wrap a HandlerFunc to be authenticated */
func (t *TokenAuth) HandleFunc(handlerFunc gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.Request
		_, err := t.Authenticate(req)
		if err != nil {
			c.JSON(200, api.MakeFailureResponse(err.Error()))
			return
		}
		handlerFunc(c)
	}
}

func (t *TokenAuth) Authenticate(req *http.Request) (*MemoryToken, error) {
	strUser := t.getter.GetUserFromRequest(req)
	if strUser == "" {
		strUser = "Anonymous"
	}
	strToken := t.getter.GetTokenFromRequest(req)
	if strToken == "" {
		return nil, errors.New("token required")
	}

	//第一次认证后，认证信息会放在内存中,过期后删除
	//首次认证需要查询zk
	token, exist, err := t.store.CheckIdToken(strUser, strToken)
	if !exist {
		//从zk获取后放入map
		zkToken, err := meta.GetUserToken(strUser)
		if err != nil {
			//避免不一致，验证失败后从内存清除
			t.store.DeleteIdToken(strUser)
			return nil, err
		}
		if zkToken != strToken {
			t.store.DeleteIdToken(strUser)
			return nil, errors.New("token invalid")
		}
		t.store.UpdateToken(strUser, strToken)
	} else {
		if err != nil {
			return nil, err
		}
	}
	return token, nil
}
