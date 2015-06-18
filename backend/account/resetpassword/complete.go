package resetpassword

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gotk/ctx"
	"github.com/gotk/pg"

	"github.com/vonwenm/go-spa/backend/account/user"
)

func init() {
	ctx.Resource("/account/reset-password/complete", &Complete{}, true)
}

type Complete struct{}

func (r *Complete) POST(c *ctx.Context, rw http.ResponseWriter, req *http.Request) error {
	db := c.Vars["db"].(*pg.Session)

	// decode request data
	var form struct {
		Password      string   `json:"password"`
		PasswordAgain string   `json:"passwordAgain"`
		ValidKey      ValidKey `json:"validKey"`
	}
	err := json.NewDecoder(req.Body).Decode(&form)
	if err != nil {
		return ctx.BadRequest(rw, c.T("reset.complete.unable_to_change"))
	}

	// validate the passwords
	if form.Password != form.PasswordAgain {
		return ctx.BadRequest(rw, c.T("reset.complete.mismatch"))
	}

	// validate the key again
	resetToken, err := getToken(db, form.ValidKey.Key)
	if err != nil || !resetToken.Valid() {
		return ctx.BadRequest(rw, c.T("reset.token.invalid_key"))
	}

	// get user from db
	u, err := user.GetById(db, resetToken.UserId)
	if err != nil {
		return ctx.InternalServerError(rw, c.T("reset.complete.user_not_found"))
	}

	// encode user password
	err = u.Password.Encode(form.Password)
	if err != nil {
		return ctx.InternalServerError(rw, c.T("reset.complete.could_not_change_password"))
	}

	// change user data in database
	err = user.Update(db, u)
	if err != nil {
		return ctx.InternalServerError(rw, c.T("reset.complete.could_not_change_password"))
	}

	// invalidate token
	err = updateToken(db, resetToken)
	if err != nil {
		log.Errorf("Unable to invalidate token: %s", err)
	}

	return ctx.OK(rw, u)
}
