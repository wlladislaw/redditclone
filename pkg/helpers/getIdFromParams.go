package helpers

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func GetOIdFromParams(r *http.Request, field string) (bson.ObjectID, error) {
	vars := mux.Vars(r)
	strId := vars[field]

	oid, err := bson.ObjectIDFromHex(strId)
	if err != nil {
		var emptyOid bson.ObjectID
		return emptyOid, err
	}
	return oid, nil
}
