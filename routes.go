package transport

import (
	"strings"

	utils "github.com/peramic/utils"
)

//TransportRoutes Transport Library Routes
var TransportRoutes = []utils.Route{

	utils.Route{
		Name:        "GetSubscribers",
		Method:      strings.ToUpper("Get"),
		Pattern:     "/rest/subscribers",
		HandlerFunc: getSubscribers,
	},
	utils.Route{
		Name:        "SetPassphrase",
		Method:      strings.ToUpper("Post"),
		Pattern:     "/rest/subscribers/certs/passphrase",
		HandlerFunc: setPassphrase,
	},
	utils.Route{
		Name:        "HasTrusted",
		Method:      strings.ToUpper("Get"),
		Pattern:     "/rest/subscribers/{id}/certs/trust",
		HandlerFunc: hasTrusted,
	},
	utils.Route{
		Name:        "DeleteTrusted",
		Method:      strings.ToUpper("Delete"),
		Pattern:     "/rest/subscribers/{id}/certs/trust",
		HandlerFunc: deleteTrusted,
	},
	utils.Route{
		Name:        "SetTrusted",
		Method:      strings.ToUpper("Post"),
		Pattern:     "/rest/subscribers/{id}/certs/trust",
		HandlerFunc: setTrusted,
	},
	utils.Route{
		Name:        "HasKeyStore",
		Method:      strings.ToUpper("Get"),
		Pattern:     "/rest/subscribers/{id}/certs/keystore",
		HandlerFunc: hasKeyStore,
	},
	utils.Route{
		Name:        "DeleteKeyStore",
		Method:      strings.ToUpper("Delete"),
		Pattern:     "/rest/subscribers/{id}/certs/keystore",
		HandlerFunc: deleteKeyStore,
	},
	utils.Route{
		Name:        "setKeyStore",
		Method:      strings.ToUpper("Post"),
		Pattern:     "/rest/subscribers/{id}/certs/keystore",
		HandlerFunc: setKeyStore,
	},
	utils.Route{
		Name:        "AddSubscriber",
		Method:      strings.ToUpper("Post"),
		Pattern:     "/rest/subscribers",
		HandlerFunc: addSubscriber,
	},
	utils.Route{
		Name:        "GetSubscriber",
		Method:      strings.ToUpper("Get"),
		Pattern:     "/rest/subscribers/{id}",
		HandlerFunc: getSubscriber,
	},
	utils.Route{
		Name:        "setSubscriber",
		Method:      strings.ToUpper("Put"),
		Pattern:     "/rest/subscribers/{id}",
		HandlerFunc: setSubscriber,
	},
	utils.Route{
		Name:        "deleteSubscriber",
		Method:      strings.ToUpper("Delete"),
		Pattern:     "/rest/subscribers/{id}",
		HandlerFunc: deleteSubscriber,
	},
}
