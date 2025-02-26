package main

import (
	"bytes"
	"encoding/json"
	"strings"
	//"fmt"
	"net/url"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Struct to capture the Salesforce token response
type AuthResponse struct {
	AccessToken string `json:"access_token"`
	InstanceURL string `json:"instance_url"`
}

type credentials struct {
	shopURL      string
	clientId     string
	clientSecret string
	tokenURL     string
	webstoreId   string
}


	//_ = tokenURL
func credential(c *gin.Context) credentials {
	var cred credentials

	// Fetch headers from the incoming request
	cred.shopURL = c.GetHeader("shopUrl")
	cred.clientId = c.GetHeader("clientId")
	cred.clientSecret = c.GetHeader("clientSecret")
	cred.webstoreId = c.GetHeader("webstoreId")

	// Validate required headers
	if cred.shopURL == "" || cred.clientId == "" || cred.clientSecret == "" || cred.webstoreId == "" {
		log.Println("Missing one or more required headers")
	}

	// Construct the token URL
	cred.tokenURL = cred.shopURL + "/services/oauth2/token?grant_type=client_credentials&client_id=" + cred.clientId + "&client_secret=" + cred.clientSecret

	return cred
}

// Function to fetch access token from Salesforce OAuth 2.0 token URL
func getAccessToken(c *gin.Context) (string, error) {
	// Make the POST request to get the token
	cred := credential(c)
	var tokenURL = cred.tokenURL

	response, err := http.Post(tokenURL, "application/x-www-form-urlencoded", nil)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	// Unmarshal the response to extract access token
	var authResponse AuthResponse
	if err := json.Unmarshal(body, &authResponse); err != nil {
		return "", err
	}

	// Return the access token
	return authResponse.AccessToken, nil
}

func getProduct(c *gin.Context) {
	productID := c.Param("id")
	getapiURL := credential(c).shopURL + "/services/data/v58.0/sobjects/Product2/" + productID

	accessToken, err := getAccessToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req, err := http.NewRequest("GET", getapiURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(response.StatusCode, gin.H{"error": err.Error()})
		return
	}

	c.Data(response.StatusCode, "application/json", body)
}

func createProduct(c *gin.Context) {
	creds := credential(c)
	postURL := creds.shopURL + "/services/data/v62.0/commerce/management/webstore/" + creds.webstoreId + "/composite-products"

	accessToken, err := getAccessToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(response.StatusCode, gin.H{"error": err.Error()})
		return
	}

	c.Data(response.StatusCode, "application/json", body)
}

func updateProduct(c *gin.Context) {
	productID := c.Param("id")
	updateURL := credential(c).shopURL + "/services/data/v58.0/sobjects/Product2/" + productID

	accessToken, err := getAccessToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req, err := http.NewRequest("PATCH", updateURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(response.StatusCode, gin.H{"error": err.Error()})
		return
	}

	c.Data(response.StatusCode, "application/json", body)
}

func deleteProduct(c *gin.Context) {
	productID := c.Param("id")
	deleteURL := credential(c).shopURL + "/services/data/v58.0/sobjects/Product2/" + productID

	accessToken, err := getAccessToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(response.StatusCode, gin.H{"error": err.Error()})
		return
	}

	c.Data(response.StatusCode, "application/json", body)
}

//Order Routes start here:

func getOrder(c *gin.Context) {
	orderID := c.Param("id")
	getapiURL := credential(c).shopURL + "/services/data/v58.0/sobjects/order/" + orderID

	accessToken, err := getAccessToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req, err := http.NewRequest("GET", getapiURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(response.StatusCode, gin.H{"error": err.Error()})
		return
	}

	c.Data(response.StatusCode, "application/json", body)
}

func createOrder(c *gin.Context) {
	creds := credential(c)
	accountID := c.Query("accountID")
	checkoutID := c.Param("checkoutId")
	postURL := creds.shopURL + "/services/data/v62.0/commerce/webstores/" + creds.webstoreId + "/checkouts/" + checkoutID + "/orders?effectiveAccountId=" + accountID

	accessToken, err := getAccessToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req, err := http.NewRequest("POST", postURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(response.StatusCode, gin.H{"error": err.Error()})
		return
	}

	c.Data(response.StatusCode, "application/json", body)
}

func updateOrder(c *gin.Context) {
	orderID := c.Param("id")
	updateURL := credential(c).shopURL + "/services/data/v58.0/sobjects/order/" + orderID

	accessToken, err := getAccessToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req, err := http.NewRequest("PATCH", updateURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(response.StatusCode, gin.H{"error": err.Error()})
		return
	}

	c.Data(response.StatusCode, "application/json", body)
}

func deleteOrder(c *gin.Context) {
	orderID := c.Param("id")
	deleteURL := credential(c).shopURL + "/services/data/v58.0/sobjects/order/" + orderID

	accessToken, err := getAccessToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(response.StatusCode, gin.H{"error": err.Error()})
		return
	}

	c.Data(response.StatusCode, "application/json", body)
}

//Account Routes start here:

func getAccount(c *gin.Context) {
	accountID := c.Param("id")
	getapiURL := credential(c).shopURL + "/services/data/v58.0/sobjects/account/" + accountID

	accessToken, err := getAccessToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req, err := http.NewRequest("GET", getapiURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(response.StatusCode, gin.H{"error": err.Error()})
		return
	}

	c.Data(response.StatusCode, "application/json", body)
}

func createAccount(c *gin.Context) {
	creds := credential(c)
	postURL := creds.shopURL + "/services/data/v58.0/sobjects/account"

	accessToken, err := getAccessToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(response.StatusCode, gin.H{"error": err.Error()})
		return
	}

	c.Data(response.StatusCode, "application/json", body)
}

func updateAccount(c *gin.Context) {
	accountID := c.Param("id")
	updateURL := credential(c).shopURL + "/services/data/v58.0/sobjects/account/" + accountID

	accessToken, err := getAccessToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req, err := http.NewRequest("PATCH", updateURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(response.StatusCode, gin.H{"error": err.Error()})
		return
	}

	c.Data(response.StatusCode, "application/json", body)
}

func deleteAccount(c *gin.Context) {
	accountID := c.Param("id")
	deleteURL := credential(c).shopURL + "/services/data/v58.0/sobjects/account/" + accountID

	accessToken, err := getAccessToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNoContent {
		c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
	} else {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			c.JSON(response.StatusCode, gin.H{"error": err.Error()})
			return
		}
		c.Data(response.StatusCode, "application/json", body)
	}
}

func getCategoryDetails(c *gin.Context) {
	name := url.QueryEscape(c.Param("name"))
	getapiURL := credential(c).shopURL + "/services/data/v58.0/query?q=select%20ID%20from%20ProductCategory%20where%20name='" + name + "'"

	accessToken, err := getAccessToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req, err := http.NewRequest("GET", getapiURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(response.StatusCode, gin.H{"error": err.Error()})
		return
	}

	c.Data(response.StatusCode, "application/json", body)
}

func getProductDetails(c *gin.Context) {
	name := url.QueryEscape(c.Param("name"))
	getapiURL := credential(c).shopURL + "/services/data/v58.0/query?q=select%20ID%20from%20Product2%20where%20name='" + name + "'"

	accessToken, err := getAccessToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req, err := http.NewRequest("GET", getapiURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(response.StatusCode, gin.H{"error": err.Error()})
		return
	}

	c.Data(response.StatusCode, "application/json", body)
}

func getPayment(c *gin.Context) {
	paymentID := c.Param("id")
	getapiURL := credential(c).shopURL + "/services/data/v58.0/sobjects/payment/" + paymentID

	accessToken, err := getAccessToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req, err := http.NewRequest("GET", getapiURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(response.StatusCode, gin.H{"error": err.Error()})
		return
	}

	c.Data(response.StatusCode, "application/json", body)
}


func createCart(c *gin.Context) {
	creds := credential(c)
	postURL := creds.shopURL + "/services/data/v62.0/commerce/webstores/" + creds.webstoreId + "/carts"
	accessToken, err := getAccessToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get access token", "details": err.Error()})
		return
	}

	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "details": err.Error()})
		return
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal JSON", "details": err.Error()})
		return
	}

	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request", "details": err.Error()})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make POST request", "details": err.Error()})
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body", "details": err.Error()})
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response", "details": string(body)})
		return
	}

	if response.StatusCode != http.StatusCreated {
		c.JSON(response.StatusCode, gin.H{"error": "Failed to create cart", "details": result})
		return
	}

	c.JSON(http.StatusCreated, result)
}


func addItemstoCart(c *gin.Context) {
	creds := credential(c)
	cartID := c.Param("cartId")
	accountID := c.Query("accountID")
	postURL := creds.shopURL + "/services/data/v62.0/commerce/webstores/" + creds.webstoreId + "/carts/" + cartID + "/cart-items?effectiveAccountId=" + accountID
	accessToken, err := getAccessToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(response.StatusCode, json.RawMessage(body))
}

func createDeliveryGroup(c *gin.Context) {
	creds := credential(c)
	cartID := c.Param("cartid")
	accountID := c.Query("accountID")

	postURL := creds.shopURL + "/services/data/v62.0/commerce/webstores/" +
		creds.webstoreId + "/carts/" + cartID +
		"/delivery-groups?effectiveAccountId=" + accountID

	accessToken, err := getAccessToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(response.StatusCode, json.RawMessage(body))
}

func createCheckout(c *gin.Context) {
	creds := credential(c)
	accountID := c.Query("accountID")

	postURL := creds.shopURL + "/services/data/v62.0/commerce/webstores/" +
		creds.webstoreId + "/checkouts?effectiveAccountId=" + accountID

	accessToken, err := getAccessToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(response.StatusCode, json.RawMessage(body))
}

func createPayment(c *gin.Context) {
	creds := credential(c)
	checkoutID := c.Param("checkoutId")
	accountID := c.Query("accountID")

	postURL := creds.shopURL + "/services/data/v62.0/commerce/webstores/" +
		creds.webstoreId + "/checkouts/" + checkoutID + "/payments?effectiveAccountId=" +
		accountID

	accessToken, err := getAccessToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(response.StatusCode, json.RawMessage(body))
}

func getOrderSummary(c *gin.Context) {
	accountID := c.Query("accountID")
	pageToken := c.Query("pageToken")
	pageSize := c.Query("pageSize")

	if accountID == "null" || accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Account ID required"})
		return
	}

	var queryParameters []string
	if pageSize != "null" && pageSize != "" {
		queryParameters = append(queryParameters, "pageSize="+pageSize)
	}
	if pageToken != "null" && pageToken != "" {
		queryParameters = append(queryParameters, "pageToken="+pageToken)
	}

	queryString := ""
	if len(queryParameters) > 0 {
		queryString = "&" + strings.Join(queryParameters, "&")
	}

	creds := credential(c)
	getapiURL := creds.shopURL + "/services/data/v62.0/commerce/webstores/" + creds.webstoreId + "/order-summaries?effectiveAccountId=" + accountID + "&ownerScoped=false&fields=AccountId&includeProducts=true" + queryString

	accessToken, err := getAccessToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req, err := http.NewRequest("GET", getapiURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(response.StatusCode, json.RawMessage(body))
}

func createCategory(c *gin.Context) {
	creds := credential(c)
	postURL := creds.shopURL + "/services/data/v58.0/sobjects/ProductCategory"
	accessToken, err := getAccessToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(response.StatusCode, json.RawMessage(body))
}

func getProductsList(c *gin.Context) {
	creds := credential(c)
	ids:=c.Query("ids")
	getapiURL := creds.shopURL + "/services/data/v62.0/commerce/webstores/"+creds.webstoreId+"/products?ids="+ids
	accessToken, err := getAccessToken(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get access token"})
		return
	}

	// Create a new request with the API URL
	req, err := http.NewRequest("GET", getapiURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// Set the Authorization header with the dynamic access token
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Send the API request
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make API request"})
		return
	}
	defer response.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}
	// Parse the JSON response
	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse JSON"})
		return
	}

	// Return the parsed JSON result
	c.JSON(http.StatusOK, result)
}

func main() {
	godotenv.Load(".env")
	router := gin.Default()

	router.Use(cors.Default())

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Salesforce Comerce Cloud Connector",
		})
	})


	//product routes
	router.GET("/getProductDetailsbyId/:id", getProduct)
	router.POST("/createProduct", createProduct)
	router.PATCH("/updateProductbyId/:id", updateProduct)
	router.DELETE("/deleteProductbyId/:id", deleteProduct)

	//order routes
	router.GET("/getOrderDetailsbyId/:id", getOrder)
	router.POST("/createOrder/:checkoutId", createOrder)
	router.PATCH("/updateOrderbyId/:id", updateOrder)
	router.DELETE("/deleteOrderbyId/:id", deleteOrder)
	router.GET("/getOrderSummary", getOrderSummary)

	//account routes
	router.GET("/getAccountDetailsbyId/:id", getAccount)
	router.POST("/createAccount", createAccount)
	router.PATCH("/updateAccountbyId/:id", updateAccount)
	router.DELETE("/deleteAccountbyId/:id", deleteAccount)

	//getCategoryId from Name
	router.GET("/getCategoryDetailsbyName/:name", getCategoryDetails)
	router.GET("/getProductDetailsbyName/:name", getProductDetails)

	//getPayment
	router.GET("/getPayment/:id", getPayment)

	//craeteCart
	router.POST("/createCart", createCart)
	router.POST("/addItemstoCart/:cartId", addItemstoCart)
	router.POST("addDeliveryGroup/:cartId", createDeliveryGroup)
	//checkoutandpayment
	router.POST("/checkout", createCheckout)
	router.POST("/setPaymentMethod/:checkoutId", createPayment)

	//additional
	router.POST("createProductCategory",createCategory)
	router.GET("listProductsbypassingIds",getProductsList)

	port := os.Getenv("CONNECTOR_ENV_PORT")
	if port == "" {
		port = "8000"
	}
	router.Run(":" + port)
}
