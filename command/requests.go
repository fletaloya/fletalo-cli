package command

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var since string

type VehicleCategory int

const (
	VehicleCategoryBici VehicleCategory = iota
	VehicleCategoryCar
	VehicleCategoryVan
	VehicleCategoryTruck
)

func init() {
	rootCmd.AddCommand(requestsCmd)
	requestsCmd.AddCommand(lastCmd)
	requestsCmd.AddCommand(listCmd)
	requestsCmd.AddCommand(showRequestCmd)
	requestsCmd.AddCommand(requestOffersCmd)
	requestsCmd.AddCommand(requestDetailCmd)
	requestsCmd.AddCommand(priceCmd)
	requestsCmd.AddCommand(requestAvailabilityCmd)
	requestsCmd.AddCommand(newRequestCmd)
	requestsCmd.PersistentFlags().StringVarP(&impersonalize, "impersonalize", "i", "me", "Run command impersonalized as other user (nickname)")
	lastCmd.PersistentFlags().StringVarP(&since, "since", "s", "1d", "Specifies timeframe to search last requests")
}

var requestsCmd = &cobra.Command{
	Use:               "requests",
	Short:             "Contains various requests subcommands",
	PersistentPreRunE: ensureAuth,
	SilenceErrors:     true,
	SilenceUsage:      true,
}

var lastCmd = &cobra.Command{
	Use:           "last",
	Short:         "Return last requests",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE:          last,
}

var listCmd = &cobra.Command{
	Use:           "list",
	Short:         "Return all requests",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE:          list,
}

var showRequestCmd = &cobra.Command{
	Use:           "show [requestID]",
	Short:         "Show request details",
	SilenceErrors: true,
	SilenceUsage:  true,
	Args:          cobra.MinimumNArgs(1),
	RunE:          showRequest,
}

var requestOffersCmd = &cobra.Command{
	Use:           "offers [requestID]",
	Short:         "Show request offers",
	SilenceErrors: true,
	SilenceUsage:  true,
	Args:          cobra.MinimumNArgs(1),
	RunE:          requestOffers,
}

var requestDetailCmd = &cobra.Command{
	Use:           "detail [requestID]",
	Short:         "Show request details",
	SilenceErrors: true,
	SilenceUsage:  true,
	Args:          cobra.MinimumNArgs(1),
	RunE:          requestDetail,
}

var priceCmd = &cobra.Command{
	Use:           "price [origin] [destination] [vehicle]",
	Short:         "Show request price",
	SilenceErrors: true,
	SilenceUsage:  true,
	PreRun:        nil,
	Args:          cobra.MinimumNArgs(3),
	RunE:          price,
}

var requestAvailabilityCmd = &cobra.Command{
	Use:           "availability [requestID]",
	Short:         "Show request availability",
	SilenceErrors: true,
	SilenceUsage:  true,
	Args:          cobra.MinimumNArgs(1),
	RunE:          requestAvailability,
}

var newRequestCmd = &cobra.Command{
	Use:           "new [requestID]",
	Short:         "Create new request",
	SilenceErrors: true,
	SilenceUsage:  true,
	Args:          cobra.MinimumNArgs(1),
	RunE:          newRequest,
}

func last(cmd *cobra.Command, args []string) error {
	url := fmt.Sprintf("%s/requests/last?since=%s&authorization=%s", getUri(), since, getToken())
	body, err := GET(url, "last requests")
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", body)

	return nil
}

func list(cmd *cobra.Command, args []string) error {
	url := fmt.Sprintf("%s/requests?authorization=%s", getUri(), getToken())
	body, err := GET(url, "requests")
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", body)

	return nil
}

func showRequest(cmd *cobra.Command, args []string) error {
	url := fmt.Sprintf("%s/request/%s?authorization=%s", getUri(), args[0], getToken())
	body, err := GET(url, "requests")
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", body)

	return nil
}

func requestOffers(cmd *cobra.Command, args []string) error {
	url := fmt.Sprintf("%s/request/%s/offers?authorization=%s", getUri(), args[0], getToken())
	body, err := GET(url, "request offers")
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", body)

	return nil
}

func requestDetail(cmd *cobra.Command, args []string) error {

	url := fmt.Sprintf("%s/request/%s?authorization=%s", getUri(), args[0], getToken())
	requestBody, err := GET(url, "requests")
	if err != nil {
		return err
	}

	url = fmt.Sprintf("%s/request/%s/remaining?authorization=%s", getUri(), args[0], getToken())
	remainingBody, err := GET(url, "remaining")
	if err != nil {
		return err
	}

	out := map[string]interface{}{}

	request := map[string]interface{}{}

	err = json.Unmarshal([]byte(requestBody), &request)
	if err != nil {
		return err
	}

	request = request["request"].(map[string]interface{})

	out["created"] = request["created"].(string)

	item := request["sections"].([]interface{})[0].(map[string]interface{})["start"].(map[string]interface{})["dropins"].([]interface{})[0].(map[string]interface{})["description"].(string)

	out["description"] = fmt.Sprintf("%s - %s", item, request["description"].(string))

	statusInt := request["status"].(float64)
	switch statusInt {
	case 0:
		out["status"] = "Pendiente"
	case 1:
		out["status"] = "Esperando respuesta"
	case 2:
		out["status"] = "Vencido"
	case 3:
		out["status"] = "Aceptado"
	case 4:
		out["status"] = "Cancelado"
	case 5:
		out["status"] = "Abortado"
	}

	remaining := map[string]interface{}{}

	err = json.Unmarshal([]byte(remainingBody), &remaining)
	if err != nil {
		return err
	}

	seconds := remaining["remaining"].(float64)
	minutes := seconds / 60

	out["remaining"] = minutes

	jsonString, err := json.Marshal(out)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", jsonString)

	return nil
}

func price(cmd *cobra.Command, args []string) error {

	_, lat1, lng1, err := latlng(args[0])
	if err != nil {
		return err
	}
	_, lat2, lng2, err := latlng(args[1])
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/route?points=%f,%f,%f,%f", getUri(), lat1, lng1, lat2, lng2)
	routeBody, err := GET(url, "route")
	if err != nil {
		return err
	}

	route := map[string]interface{}{}

	err = json.Unmarshal([]byte(routeBody), &route)
	if err != nil {
		return err
	}

	vehicle := resolveVehicle(args[2])
	distance := int(route["distance"].(float64) / 1000)

	url = fmt.Sprintf("%s/price?weight=%d&items=%d&sections=%d&vehicle=%d&distance=%d", getUri(), 1, 1, 1, vehicle, distance)
	priceBody, err := GET(url, "price")
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", priceBody)
	return nil
}

func resolveVehicle(vehicleStr string) VehicleCategory {

	var vehicle VehicleCategory

	switch vehicleStr {
	case "bici":
		vehicle = VehicleCategoryBici
	case "auto":
		vehicle = VehicleCategoryCar
	case "miniflete":
		vehicle = VehicleCategoryVan
	case "camion":
		vehicle = VehicleCategoryTruck
	}

	return vehicle
}

func requestAvailability(cmd *cobra.Command, args []string) error {

	shippers, err := requestAvailableShippers(args[0])
	if err != nil {
		return err
	}

	out := map[string]interface{}{}

	for _, value := range shippers {

		v := value.(map[string]interface{})
		commitment := int(v["commitment"].(float64) / 60)
		reputation := v["reputation"].(float64)

		shp := map[string]interface{}{"nickname": v["nickname"], "commitment": commitment, "reputation": reputation}
		out[v["id"].(string)] = shp

	}

	jsonString, err := json.Marshal(out)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", jsonString)

	return nil
}

func requestAvailableShippers(requestID string) ([]interface{}, error) {
	url := fmt.Sprintf("%s/request/%s/availability?authorization=%s", getUri(), requestID, getToken())
	availabilityBody, err := GET(url, "request availability")
	if err != nil {
		return nil, err
	}

	availability := map[string]interface{}{}

	err = json.Unmarshal([]byte(availabilityBody), &availability)
	if err != nil {
		return nil, err
	}

	shippers := availability["shippers"].([]interface{})

	return shippers, nil
}

func newRequest(cmd *cobra.Command, args []string) error {
	description := args[0]

	address1, latitude1, longitude1, err := latlng(args[1])
	if err != nil {
		return err
	}
	address2, latitude2, longitude2, err := latlng(args[2])
	if err != nil {
		return err
	}

	vehicle := resolveVehicle(args[3])

	me, err := findMe()
	if err != nil {
		return err
	}

	var senderBody map[string]interface{}
	if args[4] == "me" {
		senderBody = map[string]interface{}{"user": me["id"].(string)}
	} else {
		senderBody = map[string]interface{}{"phone": args[4]}
	}
	var receiverBody map[string]interface{}
	if args[5] == "me" {
		receiverBody = map[string]interface{}{"user": me["id"].(string)}
	} else {
		receiverBody = map[string]interface{}{"phone": args[5]}
	}

	url := fmt.Sprintf("%s/route?points=%f,%f,%f,%f", getUri(), latitude1, longitude1, latitude2, longitude2)
	body, err := GET(url, "route info")

	if err != nil {
		return err
	}

	route := map[string]interface{}{}

	err = json.Unmarshal([]byte(body), &route)
	if err != nil {
		return err
	}

	distance := route["distance"].(float64)
	sla := route["commitment"].(float64) + 1800

	return createNewRequest(description, address1, latitude1, longitude1, senderBody, address2, latitude2, longitude2, receiverBody, vehicle, sla, distance)
}

func createNewRequest(description string, addressFrom string, latitudeFrom, longitudeFrom float64, sender map[string]interface{}, addressTo string, latitudeTo, longitudeTo float64, receiver map[string]interface{}, vehicle VehicleCategory, sla, distance float64) error {

	address1 := map[string]interface{}{"address_lines": map[string]interface{}{"0": addressFrom}, "latitude": latitudeFrom, "longitude": longitudeFrom}
	address2 := map[string]interface{}{"address_lines": map[string]interface{}{"0": addressTo}, "latitude": latitudeTo, "longitude": longitudeTo}
	items := []map[string]interface{}{{"description": description, "quantity": 1, "weight": 1}}
	sections := []map[string]interface{}{map[string]interface{}{"start": map[string]interface{}{"address": address1, "player": sender, "dropins": items, "dropoffs": []map[string]interface{}{}}, "end": map[string]interface{}{"address": address2, "player": receiver, "dropins": []map[string]interface{}{}, "dropoffs": items}, "sla": sla, "distance": distance}}

	postBody := map[string]interface{}{"vehicle_category": vehicle, "sections": sections}

	url := fmt.Sprintf("%s/request?authorization=%s", getUri(), getToken())
	body, err := POST(url, postBody, "new request")

	if err != nil {
		return err
	}

	fmt.Printf("%s\v", body)

	return nil
}
