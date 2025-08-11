package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type LoadTester struct {
	client     *http.Client
	baseURL    string
	userIDs    []string
	userIDsMux sync.RWMutex
}

type User struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Age       int32  `json:"age"`
	CreatedAt int64  `json:"created_at,omitempty"`
	UpdatedAt int64  `json:"updated_at,omitempty"`
}

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int32  `json:"age"`
}

type CreateUserResponse struct {
	User User `json:"user"`
}

type GetUserResponse struct {
	User User `json:"user"`
}

type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int32  `json:"age"`
}

type UpdateUserResponse struct {
	User User `json:"user"`
}

type DeleteUserResponse struct {
	Message string `json:"message"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type ListUsersResponse struct {
	Users   []User `json:"users"`
	Total   int32  `json:"total"`
	Message string `json:"message"`
}

func NewLoadTester(baseURL string) *LoadTester {
	// Configure transport with connection pooling
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 30,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
	}

	return &LoadTester{
		client: &http.Client{
			Timeout:   30 * time.Second, // Increased timeout
			Transport: transport,
		},
		baseURL: baseURL,
		userIDs: make([]string, 0),
	}
}

// HTTP client with retry logic
func (lt *LoadTester) doRequest(req *http.Request) (*http.Response, error) {
	maxRetries := 3
	baseDelay := 100 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := lt.client.Do(req)
		if err == nil {
			return resp, nil
		}

		// Exponential backoff
		if attempt < maxRetries-1 {
			delay := baseDelay * time.Duration(1<<attempt)
			time.Sleep(delay)
		}
	}

	// Final attempt without retry
	return lt.client.Do(req)
}
func (lt *LoadTester) generateRandomUser() CreateUserRequest {
	names := []string{"Alice", "Bob", "Charlie", "Diana", "Eve", "Frank", "Grace", "Henry"}
	domains := []string{"gmail.com", "yahoo.com", "outlook.com", "company.com"}

	name := names[rand.Intn(len(names))]
	// Use timestamp and random number to ensure unique emails
	timestamp := time.Now().UnixNano()
	randomSuffix := rand.Intn(100000)
	email := fmt.Sprintf("%s.%d.%d@%s", name, timestamp, randomSuffix, domains[rand.Intn(len(domains))])
	age := int32(18 + rand.Intn(50)) // Age between 18-67

	return CreateUserRequest{
		Name:  name,
		Email: email,
		Age:   age,
	}
}

// Get random user ID from created users
func (lt *LoadTester) getRandomUserID() string {
	lt.userIDsMux.RLock()
	defer lt.userIDsMux.RUnlock()

	if len(lt.userIDs) == 0 {
		return ""
	}
	return lt.userIDs[rand.Intn(len(lt.userIDs))]
}

// Add user ID to list
func (lt *LoadTester) addUserID(id string) {
	lt.userIDsMux.Lock()
	defer lt.userIDsMux.Unlock()
	lt.userIDs = append(lt.userIDs, id)
}

// Remove user ID from list
func (lt *LoadTester) removeUserID(id string) {
	lt.userIDsMux.Lock()
	defer lt.userIDsMux.Unlock()

	for i, uid := range lt.userIDs {
		if uid == id {
			lt.userIDs = append(lt.userIDs[:i], lt.userIDs[i+1:]...)
			break
		}
	}
}

// Test CreateUser via HTTP POST /v1/users
func (lt *LoadTester) testCreateUser(ctx context.Context) error {
	req := lt.generateRandomUser()

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", lt.baseURL+"/v1/users", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := lt.doRequest(httpReq)
	if err != nil {
		return fmt.Errorf("CreateUser HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("CreateUser failed with status %d: %s", resp.StatusCode, string(body))
	}

	var createResp User
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	lt.addUserID(createResp.ID)
	log.Printf("✅ Created user: %s (ID: %s)", createResp.Name, createResp.ID)

	return nil
}

// Test GetUser via HTTP GET /v1/users/{id}
func (lt *LoadTester) testGetUser(ctx context.Context) error {
	userID := lt.getRandomUserID()
	if userID == "" {
		log.Printf("⚠️ No users available for GetUser test")
		return nil
	}

	httpReq, err := http.NewRequestWithContext(ctx, "GET", lt.baseURL+"/v1/users/"+userID, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := lt.doRequest(httpReq)
	if err != nil {
		return fmt.Errorf("GetUser HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GetUser failed with status %d: %s", resp.StatusCode, string(body))
	}

	var getUserResp User
	if err := json.NewDecoder(resp.Body).Decode(&getUserResp); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ Retrieved user: %s (ID: %s)", getUserResp.Name, getUserResp.ID)

	return nil
}

// Test UpdateUser via HTTP PUT /v1/users/{id}
func (lt *LoadTester) testUpdateUser(ctx context.Context) error {
	userID := lt.getRandomUserID()
	if userID == "" {
		log.Printf("⚠️ No users available for UpdateUser test")
		return nil
	}

	updatedData := lt.generateRandomUser()
	req := UpdateUserRequest{
		Name:  updatedData.Name + "-Updated",
		Email: updatedData.Email,
		Age:   updatedData.Age,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", lt.baseURL+"/v1/users/"+userID, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := lt.doRequest(httpReq)
	if err != nil {
		return fmt.Errorf("UpdateUser HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("UpdateUser failed with status %d: %s", resp.StatusCode, string(body))
	}

	var updateResp User
	if err := json.NewDecoder(resp.Body).Decode(&updateResp); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ Updated user: %s (ID: %s)", updateResp.Name, updateResp.ID)

	return nil
}

// Test DeleteUser via HTTP DELETE /v1/users/{id}
func (lt *LoadTester) testDeleteUser(ctx context.Context) error {
	userID := lt.getRandomUserID()
	if userID == "" {
		log.Printf("⚠️ No users available for DeleteUser test")
		return nil
	}

	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", lt.baseURL+"/v1/users/"+userID, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := lt.doRequest(httpReq)
	if err != nil {
		return fmt.Errorf("DeleteUser HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("DeleteUser failed with status %d: %s", resp.StatusCode, string(body))
	}

	var deleteResp MessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&deleteResp); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	lt.removeUserID(userID)
	log.Printf("✅ Deleted user: %s", deleteResp.Message)

	return nil
}

// Test ListUsers via HTTP GET /v1/users?page=X&limit=Y
func (lt *LoadTester) testListUsers(ctx context.Context) error {
	page := 1 + rand.Intn(3)   // Random page 1-3
	limit := 5 + rand.Intn(10) // Random limit 5-14

	url := fmt.Sprintf("%s/v1/users?page=%d&limit=%d", lt.baseURL, page, limit)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := lt.doRequest(httpReq)
	if err != nil {
		return fmt.Errorf("ListUsers HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ListUsers failed with status %d: %s", resp.StatusCode, string(body))
	}

	var listResp ListUsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("✅ Listed %d users (total: %d, page: %d)", len(listResp.Users), listResp.Total, page)

	return nil
}

// Run a single test cycle hitting all endpoints
func (lt *LoadTester) runTestCycle(ctx context.Context, cycleNum int, workerID int) {
	log.Printf("🔄 Worker %d: Starting test cycle %d", workerID, cycleNum)

	// Test all methods with some randomization
	methods := []func(context.Context) error{
		lt.testCreateUser,
		lt.testGetUser,
		lt.testUpdateUser,
		lt.testListUsers,
	}

	// Occasionally test delete (less frequent to maintain some users)
	if rand.Float32() < 0.3 { // 30% chance
		methods = append(methods, lt.testDeleteUser)
	}

	// Inline compact error scenarios (no extra funcs to keep file small)
	if rand.Float32() < 0.4 { // Not-found get
		methods = append(methods, func(c context.Context) error {
			fakeID := fmt.Sprintf("nf-%d", rand.Intn(1_000_000))
			req, _ := http.NewRequestWithContext(c, "GET", lt.baseURL+"/v1/users/"+fakeID, nil)
			resp, err := lt.doRequest(req); if err != nil { return fmt.Errorf("nf get error: %w", err) }
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusNotFound { return fmt.Errorf("intentional not found %s", fakeID) }
			return nil
		})
	}
	if rand.Float32() < 0.3 { // Invalid create
		methods = append(methods, func(c context.Context) error {
			bad := map[string]any{"name": "X", "age": -9}
			b, _ := json.Marshal(bad)
			req, _ := http.NewRequestWithContext(c, "POST", lt.baseURL+"/v1/users", bytes.NewBuffer(b))
			req.Header.Set("Content-Type", "application/json")
			resp, err := lt.doRequest(req); if err != nil { return fmt.Errorf("bad create err: %w", err) }
			defer resp.Body.Close()
			if resp.StatusCode < 400 { return fmt.Errorf("expected 4xx, got %d", resp.StatusCode) }
			return fmt.Errorf("intentional invalid create %d", resp.StatusCode)
		})
	}
	if rand.Float32() < 0.3 { // Bad list params
		methods = append(methods, func(c context.Context) error {
			req, _ := http.NewRequestWithContext(c, "GET", fmt.Sprintf("%s/v1/users?page=%s&limit=%s", lt.baseURL, "zz", "-1"), nil)
			resp, err := lt.doRequest(req); if err != nil { return fmt.Errorf("bad list err: %w", err) }
			defer resp.Body.Close()
			if resp.StatusCode < 400 { return fmt.Errorf("expected 4xx list, got %d", resp.StatusCode) }
			return fmt.Errorf("intentional bad list %d", resp.StatusCode)
		})
	}

	// Shuffle methods for variety
	rand.Shuffle(len(methods), func(i, j int) {
		methods[i], methods[j] = methods[j], methods[i]
	})

	// Execute methods with small delays
	for _, method := range methods {
		if err := method(ctx); err != nil {
			log.Printf("❌ Worker %d error in cycle %d: %v", workerID, cycleNum, err)
		}

		// Small random delay between method calls (increased for stability)
		time.Sleep(time.Duration(200+rand.Intn(600)) * time.Millisecond)
	}

	log.Printf("✅ Worker %d: Completed test cycle %d", workerID, cycleNum)
}

func main() {
	// Configuration - targeting the FastAPI client on port 8000
	baseURL := "http://rpc-client.rpc.svc.cluster.local:8000" // FastAPI client service
	interval := 2 * time.Second                               // Interval between test cycles
	numWorkers := 4                                           // Number of parallel workers

	log.Printf("🚀 Starting HTTP load tester")
	log.Printf("📡 Base URL: %s", baseURL)
	log.Printf("⏱️ Interval: %v", interval)
	log.Printf("👥 Workers: %d", numWorkers)

	// Initialize load tester
	lt := NewLoadTester(baseURL)

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Create some initial users
	log.Printf("🌱 Creating initial users...")
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		if err := lt.testCreateUser(ctx); err != nil {
			log.Printf("❌ Failed to create initial user %d: %v", i, err)
		}
		// Small delay between initial creations
		time.Sleep(200 * time.Millisecond)
	}

	// Warmup period - test service connectivity
	log.Printf("🔥 Warming up service connections...")
	for i := 0; i < 3; i++ {
		if err := lt.testListUsers(ctx); err != nil {
			log.Printf("⚠️ Warmup attempt %d failed: %v", i+1, err)
		}
		time.Sleep(500 * time.Millisecond)
	}
	log.Printf("✅ Warmup completed")

	// Main loop with parallel workers
	log.Printf("🔁 Starting continuous load testing with %d parallel workers...", numWorkers)

	// Use WaitGroup to manage workers
	var wg sync.WaitGroup

	// Start parallel workers with staggered startup
	for workerID := 1; workerID <= numWorkers; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Stagger worker startup to avoid thundering herd
			startupDelay := time.Duration(id-1) * 500 * time.Millisecond
			time.Sleep(startupDelay)
			log.Printf("🚀 Worker %d starting after %v delay", id, startupDelay)

			cycle := 1
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					lt.runTestCycle(ctx, cycle, id)
					cycle++
				}
			}
		}(workerID)
	}

	// Wait for workers (they run forever, so this will block until Ctrl+C)
	wg.Wait()
}
