package services

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	pb "github.com/appnetorg/OnlineBoutique/protos/onlineboutique"
	"github.com/appnetorg/OnlineBoutique/services/validator"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

const (
	port            = "8080"
	defaultCurrency = "CNY"
	cookieMaxAge    = 60 * 60 * 48

	cookiePrefix    = "shop_"
	cookieSessionID = cookiePrefix + "session-id"
	cookieCurrency  = cookiePrefix + "currency"
)

type ctxKeyLog struct{}
type ctxKeySessionID struct{}
type ctxKeyRequestID struct{}

type platformDetails struct {
	css      string
	provider string
}

var (
	frontendMessage  = strings.TrimSpace(os.Getenv("FRONTEND_MESSAGE"))
	isCymbalBrand    = "true" == strings.ToLower(os.Getenv("CYMBAL_BRANDING"))
	assistantEnabled = "true" == strings.ToLower(os.Getenv("ENABLE_ASSISTANT"))
	templates        = template.Must(template.New("").
				Funcs(template.FuncMap{
			"renderMoney":        renderMoney,
			"renderCurrencyLogo": renderCurrencyLogo,
		}).ParseGlob("templates/*.html"))
	plat platformDetails

	whitelistedCurrencies = map[string]bool{
		"USD": true,
		"EUR": true,
		"CAD": true,
		"JPY": true,
		"GBP": true,
		"TRY": true,
	}
)

// frontendServer implements frontendServer service
type frontendServer struct {
	port int

	productCatalogSvcAddr string
	productCatalogSvcConn *grpc.ClientConn

	currencySvcAddr string
	currencySvcConn *grpc.ClientConn

	cartSvcAddr string
	cartSvcConn *grpc.ClientConn

	recommendationSvcAddr string
	recommendationSvcConn *grpc.ClientConn

	checkoutSvcAddr string
	checkoutSvcConn *grpc.ClientConn

	shippingSvcAddr string
	shippingSvcConn *grpc.ClientConn

	adSvcAddr string
	adSvcConn *grpc.ClientConn

	shoppingAssistantSvcAddr string
}

func NewFrontendServer(port int) *frontendServer {
	return &frontendServer{
		port: port,
	}
}

// Run the server
func (fe *frontendServer) Run() error {
	mustMapEnv(&fe.productCatalogSvcAddr, "PRODUCT_CATALOG_SERVICE_ADDR")
	mustMapEnv(&fe.currencySvcAddr, "CURRENCY_SERVICE_ADDR")
	mustMapEnv(&fe.cartSvcAddr, "CART_SERVICE_ADDR")
	mustMapEnv(&fe.recommendationSvcAddr, "RECOMMENDATION_SERVICE_ADDR")
	mustMapEnv(&fe.checkoutSvcAddr, "CHECKOUT_SERVICE_ADDR")
	mustMapEnv(&fe.shippingSvcAddr, "SHIPPING_SERVICE_ADDR")
	mustMapEnv(&fe.adSvcAddr, "AD_SERVICE_ADDR")
	mustMapEnv(&fe.shoppingAssistantSvcAddr, "SHOPPING_ASSISTANT_SERVICE_ADDR")

	ctx := context.Background()
	mustConnGRPC(ctx, &fe.currencySvcConn, fe.currencySvcAddr)
	mustConnGRPC(ctx, &fe.productCatalogSvcConn, fe.productCatalogSvcAddr)
	mustConnGRPC(ctx, &fe.cartSvcConn, fe.cartSvcAddr)
	mustConnGRPC(ctx, &fe.recommendationSvcConn, fe.recommendationSvcAddr)
	mustConnGRPC(ctx, &fe.shippingSvcConn, fe.shippingSvcAddr)
	mustConnGRPC(ctx, &fe.checkoutSvcConn, fe.checkoutSvcAddr)
	mustConnGRPC(ctx, &fe.adSvcConn, fe.adSvcAddr)

	http.HandleFunc("/", fe.homeHandler)
	http.HandleFunc("/cart/checkout", fe.placeOrderHandler)
	http.HandleFunc("/cart", fe.addToCartHandler)

	log.Printf("frontendServer server running at port: %d", fe.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", fe.port), nil)
}

// homeHandler handles requests to the home page
func (fe *frontendServer) homeHandler(w http.ResponseWriter, r *http.Request) {
	userId := r.FormValue("user_id")
	log.Printf("homeHandler: Received request. UserID: %s, Currency: %s", userId, currentCurrency(r))

	// Retrieve currencies
	currencies, err := fe.getCurrencies(r.Context(), userId)
	if err != nil {
		log.Printf("homeHandler: Error retrieving currencies: %v", err)
		renderHTTPError(r, w, errors.Wrap(err, "could not retrieve currencies"), http.StatusInternalServerError)
		return
	}
	log.Printf("homeHandler: Retrieved %d currencies", len(currencies))

	// Retrieve products
	products, err := fe.getProducts(r.Context(), userId)
	if err != nil {
		log.Printf("homeHandler: Error retrieving products: %v", err)
		renderHTTPError(r, w, errors.Wrap(err, "could not retrieve products"), http.StatusInternalServerError)
		return
	}
	log.Printf("homeHandler: Retrieved %d products", len(products))

	// Retrieve cart
	cart, err := fe.getCart(r.Context(), userId)
	if err != nil {
		log.Printf("homeHandler: Error retrieving cart: %v", err)
		renderHTTPError(r, w, errors.Wrap(err, "could not retrieve cart"), http.StatusInternalServerError)
		return
	}
	log.Printf("homeHandler: Retrieved cart with %d items", cartSize(cart))

	// Process products for display
	type productView struct {
		Item  *pb.Product
		Price *pb.Money
	}
	ps := make([]productView, len(products))
	for i, p := range products {
		price, err := fe.convertCurrency(r.Context(), p.GetPriceUsd(), currentCurrency(r), userId)
		if err != nil {
			log.Printf("homeHandler: Error converting currency for product %s: %v", p.GetId(), err)
			renderHTTPError(r, w, errors.Wrapf(err, "failed to do currency conversion for product %s", p.GetId()), http.StatusInternalServerError)
			return
		}
		ps[i] = productView{p, price}
	}
	log.Printf("homeHandler: Processed %d products with currency conversion", len(ps))

	// Render template
	err = templates.ExecuteTemplate(w, "home", injectCommonTemplateData(r, map[string]interface{}{
		"show_currency": true,
		"currencies":    currencies,
		"products":      ps,
		"cart_size":     cartSize(cart),
		"banner_color":  os.Getenv("BANNER_COLOR"), // illustrates canary deployments
		"ad":            fe.chooseAd(r.Context(), []string{}, userId),
	}))
	if err != nil {
		log.Printf("homeHandler: Error rendering template: %v", err)
	}
	log.Println("homeHandler: Successfully rendered home page")
}

// placeOrderHandler handles placing an order
func (fe *frontendServer) placeOrderHandler(w http.ResponseWriter, r *http.Request) {
	// log.Println("placeOrderHandler: placing order")

	var (
		email         = r.FormValue("email")
		userId        = r.FormValue("user_id")
		streetAddress = r.FormValue("street_address")
		zipCode, _    = strconv.ParseInt(r.FormValue("zip_code"), 10, 32)
		city          = r.FormValue("city")
		state         = r.FormValue("state")
		country       = r.FormValue("country")
		ccNumber      = r.FormValue("credit_card_number")
		ccMonth, _    = strconv.ParseInt(r.FormValue("credit_card_expiration_month"), 10, 32)
		ccYear, _     = strconv.ParseInt(r.FormValue("credit_card_expiration_year"), 10, 32)
		ccCVV, _      = strconv.ParseInt(r.FormValue("credit_card_cvv"), 10, 32)
	)

	log.Printf("placeOrderHandler: received input - user_id: %s, email: %s, address: %s, city: %s, state: %s, country: %s, zip code: %d",
		userId, email, streetAddress, city, state, country, zipCode)

	payload := validator.PlaceOrderPayload{
		Email:         email,
		StreetAddress: streetAddress,
		ZipCode:       zipCode,
		City:          city,
		State:         state,
		Country:       country,
		CcNumber:      ccNumber,
		CcMonth:       ccMonth,
		CcYear:        ccYear,
		CcCVV:         ccCVV,
	}
	if err := payload.Validate(); err != nil {
		log.Printf("placeOrderHandler: validation error: %v", err)
		renderHTTPError(r, w, validator.ValidationErrorResponse(err), http.StatusUnprocessableEntity)
		return
	}
	log.Println("placeOrderHandler: input validation successful")

	order, err := pb.NewCheckoutServiceClient(fe.checkoutSvcConn).
		PlaceOrder(r.Context(), &pb.PlaceOrderRequest{
			Email: payload.Email,
			CreditCard: &pb.CreditCardInfo{
				CreditCardNumber:          payload.CcNumber,
				CreditCardExpirationMonth: int32(payload.CcMonth),
				CreditCardExpirationYear:  int32(payload.CcYear),
				CreditCardCvv:             int32(payload.CcCVV)},
			UserId:       sessionID(r),
			UserCurrency: currentCurrency(r),
			Address: &pb.Address{
				StreetAddress: payload.StreetAddress,
				City:          payload.City,
				State:         payload.State,
				ZipCode:       int32(payload.ZipCode),
				Country:       payload.Country},
		})
	if err != nil {
		log.Printf("placeOrderHandler: error placing order: %v", err)
		renderHTTPError(r, w, errors.Wrap(err, "failed to complete the order"), http.StatusInternalServerError)
		return
	}
	log.Printf("placeOrderHandler: order placed successfully, Order ID: %s", order.GetOrder().GetOrderId())

	recommendations, _ := fe.getRecommendations(r.Context(), sessionID(r), nil)
	log.Println("placeOrderHandler: retrieved recommendations")

	if len(recommendations) == 0 {
		log.Println("placeOrderHandler: No recommendations available")
	} else {
		for i, rec := range recommendations {
			log.Printf("Recommendation #%d: ID=%s, Name=%s, Description=%s, Picture=%s, PriceUSD=%v, Categories=%v",
				i+1, rec.Id, rec.Name, rec.Description, rec.Picture, rec.PriceUsd, rec.Categories)
		}
	}

	totalPaid := *order.GetOrder().GetShippingCost()
	for _, v := range order.GetOrder().GetItems() {
		multPrice := MultiplySlow(v.GetCost(), uint32(v.GetItem().GetQuantity()))
		totalPaid = *Must(Sum(&totalPaid, multPrice))
	}
	log.Printf("placeOrderHandler: total paid calculated: %d.%02d %s", totalPaid.GetUnits(), totalPaid.GetNanos()/10000000, totalPaid.GetCurrencyCode())

	currencies, err := fe.getCurrencies(r.Context(), userId)
	if err != nil {
		log.Printf("placeOrderHandler: error retrieving currencies: %v", err)
		renderHTTPError(r, w, errors.Wrap(err, "could not retrieve currencies"), http.StatusInternalServerError)
		return
	}
	log.Println("placeOrderHandler: retrieved currencies successfully")

	err = templates.ExecuteTemplate(w, "order", injectCommonTemplateData(r, map[string]interface{}{
		"show_currency":   false,
		"currencies":      currencies,
		"order":           order.GetOrder(),
		"total_paid":      &totalPaid,
		"recommendations": recommendations,
	}))
	if err != nil {
		log.Printf("placeOrderHandler: error rendering template: %v", err)
		return
	}
	log.Println("placeOrderHandler: order page rendered successfully")
}

func (fe *frontendServer) addToCartHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("addToCartHandler: Start processing request")

	quantity, _ := strconv.ParseUint(r.FormValue("quantity"), 10, 32)
	productID := r.FormValue("product_id")
	log.Printf("addToCartHandler: Received product_id=%s, quantity=%d", productID, quantity)

	payload := validator.AddToCartPayload{
		Quantity:  quantity,
		ProductID: productID,
	}

	// Validate payload
	if err := payload.Validate(); err != nil {
		log.Printf("addToCartHandler: Validation error for product_id=%s, quantity=%d: %v", productID, quantity, err)
		renderHTTPError(r, w, validator.ValidationErrorResponse(err), http.StatusUnprocessableEntity)
		return
	}
	log.Printf("addToCartHandler: Payload validated for product_id=%s, quantity=%d", productID, quantity)

	// Retrieve product details
	log.Printf("addToCartHandler: Fetching product details for product_id=%s", productID)
	p, err := fe.getProduct(r.Context(), payload.ProductID)
	if err != nil {
		log.Printf("addToCartHandler: Error retrieving product for product_id=%s: %v", productID, err)
		renderHTTPError(r, w, errors.Wrap(err, "could not retrieve product"), http.StatusInternalServerError)
		return
	}
	log.Printf("addToCartHandler: Retrieved product details for product_id=%s", productID)

	// Add product to cart
	log.Printf("addToCartHandler: Adding product_id=%s, quantity=%d to cart", productID, payload.Quantity)
	if err := fe.insertCart(r.Context(), sessionID(r), p.GetId(), int32(payload.Quantity)); err != nil {
		log.Printf("addToCartHandler: Error adding product_id=%s to cart: %v", productID, err)
		renderHTTPError(r, w, errors.Wrap(err, "failed to add to cart"), http.StatusInternalServerError)
		return
	}
	log.Printf("addToCartHandler: Successfully added product_id=%s, quantity=%d to cart", productID, payload.Quantity)

	// Redirect to cart
	w.Header().Set("location", "/cart")
	w.WriteHeader(http.StatusFound)
	log.Println("addToCartHandler: Redirected to /cart")
}

func (fe *frontendServer) getCurrencies(ctx context.Context, userID string) ([]string, error) {
	currs, err := pb.NewCurrencyServiceClient(fe.currencySvcConn).
		GetSupportedCurrencies(ctx, &pb.EmptyUser{UserId: userID})
	if err != nil {
		return nil, err
	}
	var out []string
	for _, c := range currs.CurrencyCodes {
		if _, ok := whitelistedCurrencies[c]; ok {
			out = append(out, c)
		}
	}
	return out, nil
}

func (fe *frontendServer) getProducts(ctx context.Context, userID string) ([]*pb.Product, error) {
	resp, err := pb.NewProductCatalogServiceClient(fe.productCatalogSvcConn).
		ListProducts(ctx, &pb.EmptyUser{UserId: userID})
	return resp.GetProducts(), err
}

func (fe *frontendServer) getProduct(ctx context.Context, id string) (*pb.Product, error) {
	resp, err := pb.NewProductCatalogServiceClient(fe.productCatalogSvcConn).
		GetProduct(ctx, &pb.GetProductRequest{Id: id})
	return resp, err
}

func (fe *frontendServer) getCart(ctx context.Context, userID string) ([]*pb.CartItem, error) {
	resp, err := pb.NewCartServiceClient(fe.cartSvcConn).GetCart(ctx, &pb.GetCartRequest{UserId: userID})
	return resp.GetItems(), err
}

func (fe *frontendServer) emptyCart(ctx context.Context, userID string) error {
	_, err := pb.NewCartServiceClient(fe.cartSvcConn).EmptyCart(ctx, &pb.EmptyCartRequest{UserId: userID})
	return err
}

func (fe *frontendServer) insertCart(ctx context.Context, userID, productID string, quantity int32) error {
	_, err := pb.NewCartServiceClient(fe.cartSvcConn).AddItem(ctx, &pb.AddItemRequest{
		UserId: userID,
		Item: &pb.CartItem{
			ProductId: productID,
			Quantity:  quantity},
	})
	return err
}

func (fe *frontendServer) convertCurrency(ctx context.Context, money *pb.Money, currency string, userID string) (*pb.Money, error) {
	if money.GetCurrencyCode() == currency {
		return money, nil
	}
	return pb.NewCurrencyServiceClient(fe.currencySvcConn).
		Convert(ctx, &pb.CurrencyConversionRequest{
			From:   money,
			ToCode: currency,
			UserId: userID})
}

func (fe *frontendServer) getShippingQuote(ctx context.Context, items []*pb.CartItem, currency string, userID string) (*pb.Money, error) {
	quote, err := pb.NewShippingServiceClient(fe.shippingSvcConn).GetQuote(ctx,
		&pb.GetQuoteRequest{
			Address: nil,
			Items:   items})
	if err != nil {
		return nil, err
	}
	localized, err := fe.convertCurrency(ctx, quote.GetCostUsd(), currency, userID)
	return localized, errors.Wrap(err, "failed to convert currency for shipping cost")
}

func (fe *frontendServer) getRecommendations(ctx context.Context, userID string, productIDs []string) ([]*pb.Product, error) {
	resp, err := pb.NewRecommendationServiceClient(fe.recommendationSvcConn).ListRecommendations(ctx,
		&pb.ListRecommendationsRequest{UserId: userID, ProductIds: productIDs})
	if err != nil {
		return nil, err
	}
	out := make([]*pb.Product, len(resp.GetProductIds()))
	for i, v := range resp.GetProductIds() {
		p, err := fe.getProduct(ctx, v)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get recommended product info (#%s)", v)
		}
		out[i] = p
	}
	if len(out) > 4 {
		out = out[:4] // take only first four to fit the UI
	}
	return out, err
}

func (fe *frontendServer) getAd(ctx context.Context, ctxKeys []string, userID string) ([]*pb.Ad, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*100)
	defer cancel()

	resp, err := pb.NewAdServiceClient(fe.adSvcConn).GetAds(ctx, &pb.AdRequest{
		ContextKeys: ctxKeys,
		UserId:      userID,
	})
	return resp.GetAds(), errors.Wrap(err, "failed to get ads")
}

func currentCurrency(r *http.Request) string {
	c, _ := r.Cookie(cookieCurrency)
	if c != nil {
		return c.Value
	}
	return defaultCurrency
}

func sessionID(r *http.Request) string {
	v := r.Context().Value(ctxKeySessionID{})
	if v != nil {
		return v.(string)
	}
	return ""
}

// renderHTTPError renders an error page and logs the error
func renderHTTPError(r *http.Request, w http.ResponseWriter, err error, code int) {
	log.Printf("renderHTTPError: request error: %v", err)

	errMsg := fmt.Sprintf("%+v", err)
	w.WriteHeader(code)

	// Attempt to render the error page
	templateErr := templates.ExecuteTemplate(w, "error", injectCommonTemplateData(r, map[string]interface{}{
		"error":       errMsg,
		"status_code": code,
		"status":      http.StatusText(code),
	}))
	if templateErr != nil {
		log.Printf("renderHTTPError: error rendering template: %v", templateErr)
	}
}

func renderMoney(money *pb.Money) string {
	currencyLogo := renderCurrencyLogo(money.GetCurrencyCode())
	return fmt.Sprintf("%s%d.%02d", currencyLogo, money.GetUnits(), money.GetNanos()/10000000)
}

func renderCurrencyLogo(currencyCode string) string {
	logos := map[string]string{
		"USD": "$",
		"CAD": "$",
		"JPY": "¥",
		"EUR": "€",
		"TRY": "₺",
		"GBP": "£",
	}

	logo := "$" //default
	if val, ok := logos[currencyCode]; ok {
		logo = val
	}
	return logo
}

func injectCommonTemplateData(r *http.Request, payload map[string]interface{}) map[string]interface{} {
	data := map[string]interface{}{
		"session_id":        sessionID(r),
		"request_id":        r.Context().Value(ctxKeyRequestID{}),
		"user_currency":     currentCurrency(r),
		"platform_css":      plat.css,
		"platform_name":     plat.provider,
		"is_cymbal_brand":   isCymbalBrand,
		"assistant_enabled": assistantEnabled,
		"frontendMessage":   frontendMessage,
		"currentYear":       time.Now().Year(),
	}

	for k, v := range payload {
		data[k] = v
	}

	return data
}

// get total # of items in cart
func cartSize(c []*pb.CartItem) int {
	cartSize := 0
	for _, item := range c {
		cartSize += int(item.GetQuantity())
	}
	return cartSize
}

// chooseAd queries for advertisements available and randomly chooses one, if
// available. It ignores the error retrieving the ad since it is not critical.
func (fe *frontendServer) chooseAd(ctx context.Context, ctxKeys []string, userId string) *pb.Ad {
	ads, err := fe.getAd(ctx, ctxKeys, userId)
	if err != nil {
		log.Printf("chooseAd: failed to retrieve ads: %v", err)
		return nil
	}

	// Choose a random ad from the retrieved ads
	return ads[rand.Intn(len(ads))]
}
