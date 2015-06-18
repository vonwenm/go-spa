package media

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/gotk/ctx"
	"github.com/gotk/pg"

	"github.com/vonwenm/go-spa/backend/storage/location"
	"github.com/vonwenm/go-spa/backend/storage/mediatype"
	"github.com/vonwenm/go-spa/backend/storage/mediaupload"
)

func init() {
	ctx.Resource("/media/{id:[0-9]+}", &Resource{}, false)
}

type Resource struct{}

func (r *Resource) GET(c *ctx.Context, rw http.ResponseWriter, req *http.Request) error {
	db := c.Vars["db"].(*pg.Session)
	vars := mux.Vars(req)
	id := vars["id"]

	media, err := db.FindOne(&Model{}, "id = $1", id)
	if err != nil {
		log.Errorf("Could not query media id %s: %v", id, err)
		return ctx.BadRequest(rw, c.T("media.mediaitemresource.could_not_query_media"))
	}

	return ctx.OK(rw, media)
}

func (r *Resource) PUT(c *ctx.Context, rw http.ResponseWriter, req *http.Request) error {
	db := c.Vars["db"].(*pg.Session)
	vars := mux.Vars(req)
	id := vars["id"]

	// decode request data
	var form = &MediaForm{}
	err := json.NewDecoder(req.Body).Decode(form)
	if err != nil {
		log.Errorf("Could not parse request data: %s", err)
		return ctx.BadRequest(rw, c.T("media.mediaitemresource.could_not_parse_request_data"))
	}

	// get location from database
	loc, err := location.GetById(db, form.LocationId)
	if err != nil {
		log.Errorf("Could not locate the requested location: %s", err)
		return ctx.BadRequest(rw, c.T("media.mediaitemresource.could_not_locate_requested_location"))
	}

	// get media type from database
	mediatype, err := mediatype.GetById(db, form.MediatypeId)
	if err != nil {
		log.Errorf("Could not locate the requested media type: %s", err)
		return ctx.BadRequest(rw, c.T("media.mediaitemresource.could_not_locate_requested_media_type"))
	}

	// move the uploaded file to the right place
	var dstPath string
	dstPath, err = mediaupload.MoveFile(loc, mediatype, form.Path)
	if err != nil {
		log.Errorf("Could not process the uploaded file: %s", err)
		return ctx.InternalServerError(rw, c.T("media.mediaitemresource.could_not_process_uploaded_file"))
	}

	// get media from database
	entity, err := db.FindOne(&Model{}, "id = $1", id)
	if err != nil {
		log.Errorf("Could not query media id %s: %v", id, err)
		return ctx.BadRequest(rw, c.T("media.mediaitemresource.could_not_query_media"))
	}
	media := entity.(*Model)

	// update the media
	media.Name = form.Name
	media.LocationId = form.LocationId
	media.MediatypeId = form.MediatypeId
	media.Path = dstPath
	media.EncodeData(loc, mediatype)
	err = db.Update(media)
	if err != nil {
		log.Errorf("Could not edit media %s: %v", form.Name, err)
		return ctx.BadRequest(rw, c.T("media.mediaitemresource.could_not_edit_media"))
	}

	return ctx.OK(rw, media)
}

func (r *Resource) DELETE(c *ctx.Context, rw http.ResponseWriter, req *http.Request) error {
	db := c.Vars["db"].(*pg.Session)
	vars := mux.Vars(req)
	id := vars["id"]

	media, err := db.FindOne(&Model{}, "id = $1", id)
	if err != nil {
		log.Errorf("Could not query media id %s: %v", id, err)
		return ctx.BadRequest(rw, c.T("media.mediaitemresource.could_not_query_media"))
	}
	err = db.Delete(media)
	if err != nil {
		log.Errorf("Could not delete media %s: %v", id, err)
		return ctx.InternalServerError(rw, c.T("media.mediaitemresource.could_not_delete_media"))
	}
	return ctx.NoContent(rw)
}
