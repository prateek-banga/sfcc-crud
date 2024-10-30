package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	//"log"
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
	shopURL       string
	clientID      string
	client_secret string
	tokenURL      string
	webstoreId    string
}

func credential() credentials {
	var cred credentials
	cred.shopURL = os.Getenv("shop_url")
	cred.clientID = os.Getenv("clientID")
	cred.client_secret = os.Getenv("clientSecret")
	cred.webstoreId = os.Getenv("webstoreId")
	cred.tokenURL = cred.shopURL + "/services/oauth2/token?grant_type=client_credentials&client_id=" + cred.clientID + "&client_secret=" + cred.client_secret
	return cred
	//_ = tokenURL

}

// Function to fetch access token from Salesforce OAuth 2.0 token URL
func getAccessToken() (string, error) {
	// Make the POST request to get the token
	var credentials_for_access_token credentials = credential()
	//credentials_for_access_token = credential()
	var tokenURL = credentials_for_access_token.tokenURL

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
	getapiURL := credential().shopURL + "/services/data/v57.0/commerce/webstores/" + credential().webstoreId + "/products/" + productID

	// Get the dynamic access token
	accessToken, err := getAccessToken()
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

// Function to create a product and associate it with a category
func createProduct(c *gin.Context) {
	creds := credential()
	postURL := creds.shopURL + "/services/data/v58.0/sobjects/Product2"
	accessToken, err := getAccessToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get access token"})
		return
	}

	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal JSON"})
		return
	}

	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make POST request"})
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}

	productID, ok := result["id"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Product ID"})
		return
	}

	myMap, err := associateProductWithCategory(productID, os.Getenv("categoryID"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to associate product with category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product associated with category successfully",
		"data":    myMap,
	})
}

// Function to associate a product with a category in Salesforce
func associateProductWithCategory(productID string, categoryID string) (map[string]interface{}, error) {
	creds := credential()
	postCategoryProductURL := creds.shopURL + "/services/data/v58.0/sobjects/ProductCategoryProduct"
	accessToken, err := getAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token")
	}

	// Prepare the request body
	categoryProductData := map[string]interface{}{
		"ProductCategoryId": categoryID,
		"ProductId":         productID,
	}

	jsonCategoryProductData, err := json.Marshal(categoryProductData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal category product JSON")
	}

	req, err := http.NewRequest("POST", postCategoryProductURL, bytes.NewBuffer(jsonCategoryProductData))
	if err != nil {
		return nil, fmt.Errorf("failed to create category product request")
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make POST request to associate product with category")
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read category product response body")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse category product response")
	}

	return result, nil
}

func updateProduct(c *gin.Context) {
	productID := c.Param("id")
	updateURL := credential().shopURL + "/services/data/v58.0/sobjects/Product2/" + productID

	accessToken, err := getAccessToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get access token"})
		return
	}

	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal JSON"})
		return
	}

	req, err := http.NewRequest("PATCH", updateURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make PATCH request"})
		return
	}
	defer response.Body.Close()

	// body, err := ioutil.ReadAll(response.Body)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
	// 	return
	// }

	// var result interface{}
	// if err := json.Unmarshal(body, &result); err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
	// 	return
	// }

	// c.JSON(http.StatusOK, gin.H{"response": result})
	c.JSON(http.StatusOK, gin.H{"response": "Product updated"})
}

func deleteProduct(c *gin.Context) {
	productID := c.Param("id")
	deleteURL := credential().shopURL + "/services/data/v58.0/sobjects/Product2/" + productID

	accessToken, err := getAccessToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get access token"})
		return
	}

	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make DELETE request"})
		return
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNoContent {
		c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
	}
}

//Order Routes start here:

func getOrder(c *gin.Context) {
	orderID := c.Param("id")
	getapiURL := credential().shopURL + "/services/data/v58.0/sobjects/order/" + orderID

	// Get the dynamic access token
	accessToken, err := getAccessToken()
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

func createOrder(c *gin.Context) {
	creds := credential()
	postURL := creds.shopURL + "/services/data/v58.0/sobjects/order"
	accessToken, err := getAccessToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get access token"})
		return
	}

	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal JSON"})
		return
	}

	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make POST request"})
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Order created successfully",
	})
}

func updateOrder(c *gin.Context) {
	orderID := c.Param("id")
	updateURL := credential().shopURL + "/services/data/v58.0/sobjects/order/" + orderID

	accessToken, err := getAccessToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get access token"})
		return
	}

	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal JSON"})
		return
	}

	req, err := http.NewRequest("PATCH", updateURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make PATCH request"})
		return
	}
	defer response.Body.Close()

	c.JSON(http.StatusOK, gin.H{"response": "Order updated"})
}
func deleteOrder(c *gin.Context) {
	orderID := c.Param("id")
	deleteURL := credential().shopURL + "/services/data/v58.0/sobjects/order/" + orderID

	accessToken, err := getAccessToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get access token"})
		return
	}

	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make DELETE request"})
		return
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNoContent {
		c.JSON(http.StatusOK, gin.H{"message": "Order deleted successfully"})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order"})
	}
}

//Customer Routes

func getCustomer(c *gin.Context) {
	customerID := c.Param("id")
	getapiURL := credential().shopURL + "/services/data/v58.0/sobjects/customer/" + customerID

	// Get the dynamic access token
	accessToken, err := getAccessToken()
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

func createCustomer(c *gin.Context) {
	creds := credential()
	postURL := creds.shopURL + "/services/data/v58.0/sobjects/customer"
	accessToken, err := getAccessToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get access token"})
		return
	}

	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal JSON"})
		return
	}

	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make POST request"})
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Customer created successfully",
	})
}

func updateCustomer(c *gin.Context) {
	customerID := c.Param("id")
	updateURL := credential().shopURL + "/services/data/v58.0/sobjects/customer/" + customerID

	accessToken, err := getAccessToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get access token"})
		return
	}

	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal JSON"})
		return
	}

	req, err := http.NewRequest("PATCH", updateURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make PATCH request"})
		return
	}
	defer response.Body.Close()

	c.JSON(http.StatusOK, gin.H{"response": "Customer updated"})
}
func deleteCustomer(c *gin.Context) {
	orderID := c.Param("id")
	deleteURL := credential().shopURL + "/services/data/v58.0/sobjects/order/" + orderID

	accessToken, err := getAccessToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get access token"})
		return
	}

	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make DELETE request"})
		return
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNoContent {
		c.JSON(http.StatusOK, gin.H{"message": "Customer deleted successfully"})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order"})
	}
}

//Account Routes start here:

func getAccount(c *gin.Context) {
	accountID := c.Param("id")
	getapiURL := credential().shopURL + "/services/data/v58.0/sobjects/account/" + accountID

	// Get the dynamic access token
	accessToken, err := getAccessToken()
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

func createAccount(c *gin.Context) {
	creds := credential()
	postURL := creds.shopURL + "/services/data/v58.0/sobjects/account"
	accessToken, err := getAccessToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get access token"})
		return
	}

	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal JSON"})
		return
	}

	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make POST request"})
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Account created successfully",
	})
}

func updateAccount(c *gin.Context) {
	accountID := c.Param("id")
	updateURL := credential().shopURL + "/services/data/v58.0/sobjects/account/" + accountID

	accessToken, err := getAccessToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get access token"})
		return
	}

	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal JSON"})
		return
	}

	req, err := http.NewRequest("PATCH", updateURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make PATCH request"})
		return
	}
	defer response.Body.Close()

	c.JSON(http.StatusOK, gin.H{"response": "Account updated"})
}
func deleteAccount(c *gin.Context) {
	accountID := c.Param("id")
	deleteURL := credential().shopURL + "/services/data/v58.0/sobjects/account/" + accountID

	accessToken, err := getAccessToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get access token"})
		return
	}

	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to make DELETE request"})
		return
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNoContent {
		c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
	}
}

func getCategoryDetails(c *gin.Context) {
	name := c.Param("name")
	getapiURL := credential().shopURL + "/services/data/v58.0/query?q=select ID from ProductCategory where name='" + name + "'"
	// Get the dynamic access token
	accessToken, err := getAccessToken()
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
	//fmt.Println(credential().tokenURL)
	router := gin.Default()

	router.Use(cors.Default())

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Salesforce Comerce Cloud Connector",
		})
	})

	//product routes
	router.GET("/getProduct/:id", getProduct)
	router.POST("/createProduct", createProduct)
	router.PATCH("/updateProduct/:id", updateProduct)
	router.DELETE("/deleteProduct/:id", deleteProduct)

	//order routes
	router.GET("/getOrder/:id", getOrder)
	router.POST("/createOrder", createOrder)
	router.PATCH("/updateOrder/:id", updateOrder)
	router.DELETE("/deleteOrder/:id", deleteOrder)

	//customer routes
	router.GET("/getCustomer/:id", getCustomer)
	router.POST("/createCustomer", createCustomer)
	router.PATCH("/updateCustomer/:id", updateCustomer)
	router.DELETE("/deleteCustomer/:id", deleteCustomer)

	//account routes
	router.GET("/getAccount/:id", getAccount)
	router.POST("/createAccount", createAccount)
	router.PATCH("/updateAccount/:id", updateAccount)
	router.DELETE("/deleteAccount/:id", deleteAccount)

	//getCategoryId from Name
	router.GET("getCategoryDetails/:name", getCategoryDetails)

	port := os.Getenv("CONNECTOR_ENV_PORT")
	if port == "" {
		port = "8000"
	}
	router.Run(":" + port)
}
