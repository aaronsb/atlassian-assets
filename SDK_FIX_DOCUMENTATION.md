# go-atlassian SDK Filter Method Fix

## Issue Summary

The go-atlassian SDK v2.6.1 has a critical bug in the `Object.Filter()` method that makes AQL (Assets Query Language) searches completely non-functional. This affects both search and list operations that rely on AQL queries.

## Problem Description

### Symptoms
- All AQL queries return 0 results despite objects existing
- API returns 200 OK status but empty result sets
- No error messages - appears to work but returns no data
- Affects both `SearchObjects()` and `ListObjects()` methods

### Root Cause Analysis
The issue is in the go-atlassian SDK's Filter implementation at:
```
/go/pkg/mod/github.com/ctreminiom/go-atlassian/v2@v2.6.1/assets/internal/object_impl.go:148-180
```

**Evidence that proves SDK is broken:**
1. **Direct curl calls work perfectly** with identical auth and queries
2. **Other SDK methods work fine** (schemas, object types, etc.)
3. **API returns 200 OK** but SDK returns empty objects
4. **AQL syntax is correct** - verified against official documentation

### Test Case That Exposed the Bug
```bash
# Working direct curl call
curl -u email:token \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"qlQuery": "objectSchemaId = 3"}' \
  "https://api.atlassian.com/jsm/assets/workspace/WORKSPACE_ID/v1/object/aql"
# Returns objects successfully

# Broken SDK call
objects, response, err := client.Object.Filter(ctx, workspaceID, "objectSchemaId = 3", true, 50, 0)
// Returns: err=nil, response.Code=200, objects.Total=0 (but objects exist!)
```

## Solution Implementation

### Direct HTTP Replacement Method
Created `searchObjectsDirect()` in `internal/client/client.go` that bypasses the broken SDK:

```go
func (ac *AssetsClient) searchObjectsDirect(ctx context.Context, aqlQuery string, maxResults int) (*models.ObjectListResultScheme, error) {
    // Build endpoint URL using API gateway (same as SDK)
    endpoint := fmt.Sprintf("https://api.atlassian.com/jsm/assets/workspace/%s/v1/object/aql", 
        ac.workspaceID)
    
    // Create request payload with all parameters
    payload := map[string]interface{}{
        "qlQuery":           aqlQuery,
        "startAt":           0,
        "maxResults":        maxResults,
        "includeAttributes": true,
    }
    
    payloadBytes, err := json.Marshal(payload)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal payload: %w", err)
    }
    
    // Create HTTP request
    req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(payloadBytes))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    // Set headers
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "application/json")
    
    // Set Basic Auth
    auth := base64.StdEncoding.EncodeToString([]byte(ac.config.GetUsername() + ":" + ac.config.GetPassword()))
    req.Header.Set("Authorization", "Basic " + auth)
    
    // Make the request
    resp, err := ac.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to execute request: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != 200 {
        var errorBody bytes.Buffer
        errorBody.ReadFrom(resp.Body)
        return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, errorBody.String())
    }
    
    // Parse response
    var result models.ObjectListResultScheme
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    return &result, nil
}
```

### Modified Methods to Use Direct HTTP
Updated both `SearchObjects()` and `ListObjects()` methods to use the direct HTTP implementation:

```go
// SearchObjects now calls searchObjectsDirect instead of SDK Filter
func (ac *AssetsClient) SearchObjects(ctx context.Context, query string, limit int) (*Response, error) {
    objects, err := ac.searchObjectsDirect(ctx, query, limit)
    if err != nil {
        return NewErrorResponse(fmt.Errorf("failed to search objects: %w", err)), nil
    }
    
    return NewSuccessResponse(map[string]interface{}{
        "objects": objects.Values,
        "total":   objects.Total,
        "query":   query,
    }), nil
}

// ListObjects now calls searchObjectsDirect with schema filter
func (ac *AssetsClient) ListObjects(ctx context.Context, schemaID string, limit int) (*Response, error) {
    query := fmt.Sprintf("objectSchemaId = %s", schemaID)
    objects, err := ac.searchObjectsDirect(ctx, query, limit)
    // ... rest of implementation
}
```

## Key Technical Details

### Authentication
- Uses `ac.config.GetUsername()` (email) and `ac.config.GetPassword()` (API token)
- Implements Base64-encoded Basic Auth: `base64(email:token)`
- Same authentication as working SDK methods

### API Endpoint
- **Correct URL**: `https://api.atlassian.com/jsm/assets/workspace/{workspaceID}/v1/object/aql`
- **Not customer instance URL**: `https://customer.atlassian.net/jira/assets/...` (returns HTML)
- Uses API gateway for proper routing and authentication

### Request Format
- **Method**: POST (not GET)
- **Content-Type**: `application/json`
- **Body**: JSON payload with all parameters (not query string)
- **Parameters**: `qlQuery`, `startAt`, `maxResults`, `includeAttributes`

### Response Handling
- Expects JSON response in `models.ObjectListResultScheme` format
- Proper error handling for non-200 status codes
- JSON decoding with detailed error messages

## Validation Results

### Before Fix
```bash
$ ./assets search --query "objectSchemaId = 3"
{
  "data": {
    "objects": [],
    "total": 0,
    "query": "objectSchemaId = 3"
  },
  "success": true
}
```

### After Fix
```bash
$ ./assets search --query "objectSchemaId = 3"
DEBUG: SearchObjects direct HTTP call successful - found 8 objects
{
  "data": {
    "objects": [
      {
        "workspaceId": "d683300e-ec06-45ee-8789-b0f5e219c16f",
        "globalId": "d683300e-ec06-45ee-8789-b0f5e219c16f:1020",
        "id": "1020",
        "label": "Blue Barn #2",
        "objectKey": "COMPUTE-1020",
        // ... full object data
      },
      // ... 7 more objects
    ],
    "total": 8,
    "query": "objectSchemaId = 3"
  },
  "success": true
}
```

## Files Modified

1. **`internal/client/client.go`** (Lines 374-443)
   - Added `searchObjectsDirect()` method
   - Modified `SearchObjects()` to use direct HTTP
   - Modified `ListObjects()` to use direct HTTP
   - Added required imports: `bytes`, `encoding/base64`

## Future Considerations

### Contributing Back to SDK
**GitHub Issue Created**: https://github.com/ctreminiom/go-atlassian/issues/387

We've reported this bug to the upstream go-atlassian repository with:
- Detailed reproduction steps and code samples
- Root cause analysis proving it's an SDK bug
- Working workaround implementation
- Offer to submit a pull request with the fix

**Next steps for SDK contribution:**
1. **Wait for maintainer response** on issue #387
2. **Clone the original SDK** if pull request is welcomed: `git clone https://github.com/ctreminiom/go-atlassian.git`
3. **Identify the exact bug** in `assets/internal/object_impl.go:148-180`
4. **Create a working fix** based on our direct HTTP implementation
5. **Submit pull request** with test cases that demonstrate the fix

**User's Fork**: https://github.com/aaronsb/go-atlassian (for potential PR development)

### Potential SDK Bug
The issue may be in how the SDK constructs the HTTP request or processes the response. Common possibilities:
- Incorrect Content-Type header
- Malformed JSON payload structure
- Response parsing/unmarshaling issues
- Incorrect parameter encoding

## Testing Commands

### Test Search Functionality
```bash
# Test basic search
./assets search --query "objectSchemaId = 3"

# Test object type search
./assets search --query "objectTypeId = 156"

# Test name search
./assets search --query "Name like 'Blue%'"
```

### Test List Functionality
```bash
# Test schema listing
./assets list --schema 3

# Test different schema
./assets list --schema 2
```

### Debug Mode
Both commands include debug logging that shows:
- Workspace ID being used
- Exact HTTP endpoint URL
- Request payload sent
- HTTP response status
- Number of objects found

## Related Documentation

- **AQL Documentation**: https://support.atlassian.com/assets/docs/use-assets-query-language-aql/
- **SDK Documentation**: https://docs.go-atlassian.io/jira-assets/aql
- **Test Plan**: See `tests.md` for comprehensive test coverage

## Impact

This fix restores critical search and list functionality that was completely broken due to the SDK bug. Both `assets search` and `assets list` commands now work reliably and return complete object data as expected.

## Additional AQL Limitation Discovered

During implementation of the dual search functionality, we discovered that **AQL LIKE queries are non-functional** in the current Atlassian Assets environment. This affects partial matching capabilities:

### What Works:
- Exact matches: `Name = "Blue Barn #2"`
- Schema/Type filters: `objectSchemaId = 3`
- Status filters: `Status = "Active"`
- Wildcard queries using inequality: `Name != ""`

### What Doesn't Work:
- Partial matches: `Name like "%Blue%"`
- Starts with: `Name like "Blue%"`
- Ends with: `Name like "%Barn"`
- Case-insensitive searches with LIKE

### Impact on Simple Search:
The simple search functionality has been adapted to work within these constraints:
- `=exact` - Exact match (works)
- `^exact$` - Exact match with anchors (works)
- `*` - Match all objects (works using `Name != ""`)
- Partial patterns - **Not supported** due to LIKE limitation

This limitation affects user experience but exact matching still provides useful search capabilities for known object names and keys.