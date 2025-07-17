# Atlassian Assets CLI - Comprehensive Test Plan

## Test Strategy

**Approach:** Live environment testing with dedicated test schema
**Coverage:** All 21 CLI commands with subcommands and contextual hints validation
**Environment:** Real Atlassian Assets workspace with test data isolation

## Test Infrastructure

### Test Schema Setup
```bash
# Create isolated test environment
assets test create-schema --with-sample-data --name "CLI_Test_Schema_$(date +%Y%m%d)"

# Verify test schema creation
assets schema list | grep CLI_Test_Schema
```

### Test Data Requirements
- **Test Schema:** Dedicated schema for testing (auto-generated name)
- **Sample Object Types:** Servers, Applications, Locations (created by test setup)
- **Sample Objects:** Minimal test instances for each type
- **Cleanup:** Automated test schema removal after testing

## Test Categories

### 1. Core CRUD Operations
**Commands:** `create`, `list`, `get`, `update`, `delete`, `search`

### 2. Schema Management  
**Commands:** `schema`, `attributes`, `validate`

### 3. Advanced Workflows
**Commands:** `browse`, `extract`, `apply`, `catalog`, `trace`

### 4. Intelligence Features
**Commands:** `complete`, `workflows`, `summary`

### 5. Utility Commands
**Commands:** `config`, `resolve`, `copy-attributes`, `test`

### 6. System Commands
**Commands:** `help`, `completion`

---

## Individual Command Tests

### Test 01: `apply` - Apply attributes to object types and objects

**Test Cases:**
- [ ] **T01.1** - Apply attributes to object type
  ```bash
  assets apply attributes --help
  assets apply attributes --to-object-type {test_type} --attributes-file {test_file}
  ```
- [ ] **T01.2** - Validate contextual hints for apply operations
- [ ] **T01.3** - Error handling for invalid attribute files
- [ ] **T01.4** - Success flow with proper attribute application

**Expected Results:**
- Command returns success with applied attributes
- Contextual hints suggest next steps (create instances, validation)
- Error handling provides helpful guidance

---

### Test 02: `attributes` - Get object type attributes

**Test Cases:**
- [ ] **T02.1** - Get attributes by object type ID
  ```bash
  assets attributes --help
  assets attributes --type {test_object_type_id}
  ```
- [ ] **T02.2** - Get attributes with schema resolution
- [ ] **T02.3** - Error handling for non-existent object types
- [ ] **T02.4** - Validate attribute details and structure

**Expected Results:**
- Returns complete attribute schema including data types, constraints
- Shows required vs optional attributes
- Includes reference information for reference attributes

---

### Test 03: `browse` - High-level browsing and exploration tools

**Test Cases:**
- [ ] **T03.1** - Browse hierarchy in test schema
  ```bash
  assets browse --help
  assets browse hierarchy --schema {test_schema_id}
  ```
- [ ] **T03.2** - Browse children of parent object type
  ```bash
  assets browse children --parent {parent_id} --schema {test_schema_id}
  ```
- [ ] **T03.3** - Browse attributes comparison
  ```bash
  assets browse attrs --source {source_id} --target {target_id}
  ```
- [ ] **T03.4** - Validate hierarchical relationships display

**Expected Results:**
- Clear visualization of object type hierarchies
- Parent-child relationships correctly displayed
- Attribute comparison shows differences and similarities
- Navigation hints for deeper exploration

---

### Test 04: `catalog` - Browse global catalogs

**Test Cases:**
- [ ] **T04.1** - Catalog all attributes across workspace
  ```bash
  assets catalog --help
  assets catalog attributes
  ```
- [ ] **T04.2** - Search catalog with pattern matching
  ```bash
  assets catalog attributes --pattern "name|title|label"
  ```
- [ ] **T04.3** - Catalog with pagination
  ```bash
  assets catalog attributes --page 2 --per-page 25
  ```
- [ ] **T04.4** - Schema-specific catalog browsing
  ```bash
  assets catalog attributes --schema {test_schema_id}
  ```

**Expected Results:**
- Global attribute inventory with search capabilities
- Pattern matching finds relevant attributes
- Pagination works correctly for large result sets
- Results include attribute metadata (type, references, requirements)

---

### Test 05: `complete` - Intelligently complete object properties

**Test Cases:**
- [ ] **T05.1** - Complete object with minimal input
  ```bash
  assets complete --help
  assets complete --type {test_object_type} --data '{"name":"Test Item"}'
  ```
- [ ] **T05.2** - Complete with validation errors
- [ ] **T05.3** - Complete with reference attributes
- [ ] **T05.4** - Validate completion suggestions and defaults

**Expected Results:**
- Intelligent completion fills missing required fields
- Provides suggestions for optional fields
- Validates input against schema constraints
- Returns actionable completion recommendations

---

### Test 06: `completion` - Generate shell autocompletion

**Test Cases:**
- [ ] **T06.1** - Generate bash completion
  ```bash
  assets completion --help
  assets completion bash
  ```
- [ ] **T06.2** - Generate zsh completion
- [ ] **T06.3** - Validate completion script syntax

**Expected Results:**
- Valid shell completion scripts generated
- Scripts provide command and flag completion
- No syntax errors in generated scripts

---

### Test 07: `config` - Manage configuration

**Test Cases:**
- [ ] **T07.1** - Show current configuration
  ```bash
  assets config --help
  assets config show
  ```
- [ ] **T07.2** - Test connection
  ```bash
  assets config test
  ```
- [ ] **T07.3** - Discover workspace
  ```bash
  assets config discover-workspace
  ```

**Expected Results:**
- Configuration displayed with sensitive data masked
- Connection test validates API access
- Workspace discovery finds available workspaces

---

### Test 08: `copy-attributes` - Copy attributes between object types

**Test Cases:**
- [ ] **T08.1** - Copy attributes from source to target
  ```bash
  assets copy-attributes --help
  assets copy-attributes --from {source_type} --to {target_type}
  ```
- [ ] **T08.2** - Handle reference attribute copying
- [ ] **T08.3** - Validate attribute compatibility

**Expected Results:**
- Successful attribute copying between compatible types
- Reference attributes handled appropriately
- Clear feedback on copy operations and any conflicts

---

### Test 09: `create` - Create assets with guided workflow

**Test Cases:**
- [ ] **T09.1** - Create object type with guided flow
  ```bash
  assets create --help
  assets create object-type --schema {test_schema} --name "Test Servers"
  ```
- [ ] **T09.2** - Create instance (legacy)
  ```bash
  assets create instance --schema {test_schema} --type {test_type} --data '{"name":"SERVER-001"}'
  ```
- [ ] **T09.3** - Validate contextual hints after creation

**Expected Results:**
- Object types created successfully with proper metadata
- Contextual hints guide next steps (add attributes, create children)
- Legacy instance creation works but suggests better approach

---

### Test 10: `delete` - Delete an asset object

**Test Cases:**
- [ ] **T10.1** - Delete specific object
  ```bash
  assets delete --help
  assets delete --id {test_object_id}
  ```
- [ ] **T10.2** - Handle non-existent object deletion
- [ ] **T10.3** - Validate deletion confirmation

**Expected Results:**
- Objects deleted successfully when they exist
- Appropriate error handling for non-existent objects
- Clear feedback on deletion operations

---

### Test 11: `extract` - Extract attributes from objects and object types

**Test Cases:**
- [ ] **T11.1** - Extract attributes from object type
  ```bash
  assets extract --help
  assets extract attributes --from-object-type {test_type}
  ```
- [ ] **T11.2** - Extract with output file
- [ ] **T11.3** - Validate extracted attribute format

**Expected Results:**
- Attributes extracted in reusable format
- Output suitable for apply operations
- Includes all necessary attribute metadata

---

### Test 12: `get` - Get a specific asset object

**Test Cases:**
- [ ] **T12.1** - Get object by ID
  ```bash
  assets get --help
  assets get --id {test_object_id}
  ```
- [ ] **T12.2** - Handle non-existent object retrieval
- [ ] **T12.3** - Validate object details completeness

**Expected Results:**
- Complete object details returned including attributes
- Error handling for invalid or non-existent IDs
- Structured data suitable for further processing

---

### Test 13: `help` - Help about any command

**Test Cases:**
- [ ] **T13.1** - General help display
  ```bash
  assets help
  ```
- [ ] **T13.2** - Command-specific help
  ```bash
  assets help create
  ```
- [ ] **T13.3** - Subcommand help
  ```bash
  assets help create object-type
  ```

**Expected Results:**
- Comprehensive help documentation
- Accurate usage examples
- Consistent formatting across all commands

---

### Test 14: `list` - List asset objects

**Test Cases:**
- [ ] **T14.1** - List objects in schema (fixed bug validation)
  ```bash
  assets list --help
  assets list --schema {test_schema}
  ```
- [ ] **T14.2** - List with type filter
  ```bash
  assets list --schema {test_schema} --type {test_type}
  ```
- [ ] **T14.3** - List with AQL filter
  ```bash
  assets list --schema {test_schema} --filter "Name like 'Test%'"
  ```

**Expected Results:**
- Objects listed successfully (no 404 errors after bug fix)
- Filters work correctly to narrow results
- Pagination handled appropriately for large result sets

---

### Test 15: `resolve` - Resolve between human names and internal IDs

**Test Cases:**
- [ ] **T15.1** - Resolve schema names to IDs
  ```bash
  assets resolve --help
  assets resolve schema --name {test_schema_name}
  ```
- [ ] **T15.2** - Resolve object type names
  ```bash
  assets resolve type --name {test_type_name} --schema {test_schema}
  ```
- [ ] **T15.3** - Get resolver statistics
  ```bash
  assets resolve stats
  ```

**Expected Results:**
- Name-to-ID resolution works accurately
- Reverse ID-to-name resolution supported
- Resolver statistics show cache performance

---

### Test 16: `schema` - Manage asset schemas

**Test Cases:**
- [ ] **T16.1** - List all schemas
  ```bash
  assets schema --help
  assets schema list
  ```
- [ ] **T16.2** - Get schema details
  ```bash
  assets schema get --id {test_schema}
  ```
- [ ] **T16.3** - List object types in schema
  ```bash
  assets schema types --schema {test_schema}
  ```
- [ ] **T16.4** - Create object type in schema
  ```bash
  assets schema create-type --schema {test_schema} --name "Test Type"
  ```

**Expected Results:**
- Schema listing shows all available schemas
- Schema details include metadata and statistics
- Object type operations work within schema context

---

### Test 17: `search` - Search for asset objects

**Test Cases:**
- [ ] **T17.1** - Basic AQL search
  ```bash
  assets search --help
  assets search --query "objectTypeId = {test_type_id}"
  ```
- [ ] **T17.2** - Complex AQL search with conditions
  ```bash
  assets search --query "Name like 'Test%' AND objectSchemaId = {test_schema}"
  ```
- [ ] **T17.3** - Search result pagination

**Expected Results:**
- AQL queries execute correctly
- Results match search criteria
- Complex queries with multiple conditions work
- Search performance acceptable for reasonable data sets

---

### Test 18: `summary` - High-level analysis and summary tools

**Test Cases:**
- [ ] **T18.1** - Completion summary analysis
  ```bash
  assets summary --help
  assets summary completion --type {test_type} --data '{"name":"test"}'
  ```
- [ ] **T18.2** - Schema summary statistics
  ```bash
  assets summary schema --id {test_schema}
  ```

**Expected Results:**
- Completion summaries provide actionable insights
- Schema summaries show comprehensive statistics
- Analysis helps understand data structure and gaps

---

### Test 19: `test` - Test environment setup and validation tools

**Test Cases:**
- [ ] **T19.1** - Create test schema
  ```bash
  assets test --help
  assets test create-schema --with-sample-data
  ```
- [ ] **T19.2** - Cleanup test environments
  ```bash
  assets test cleanup --prefix TEST --dry-run
  ```

**Expected Results:**
- Test schemas created with predictable structure
- Sample data populated appropriately
- Cleanup operations identify test schemas correctly

---

### Test 20: `trace` - Trace and discover references across schemas

**Test Cases:**
- [ ] **T20.1** - Trace attribute references
  ```bash
  assets trace --help
  assets trace reference --attribute-id {test_attr_id}
  ```
- [ ] **T20.2** - Trace object type dependencies
  ```bash
  assets trace dependencies --object-type {test_type} --schema {test_schema}
  ```

**Expected Results:**
- Reference tracing discovers cross-schema links
- Dependency analysis shows complete dependency trees
- Reference resolution handles complex relationship chains

---

### Test 21: `update` - Update an existing asset object

**Test Cases:**
- [ ] **T21.1** - Update object properties
  ```bash
  assets update --help
  assets update --id {test_object_id} --data '{"name":"Updated Name"}'
  ```
- [ ] **T21.2** - Handle invalid update data
- [ ] **T21.3** - Validate update success

**Expected Results:**
- Object updates applied correctly
- Validation prevents invalid updates
- Clear feedback on update operations

---

### Test 22: `validate` - Validate object properties against schema

**Test Cases:**
- [ ] **T22.1** - Validate object properties
  ```bash
  assets validate --help
  assets validate --type {test_type} --data '{"name":"Test Item"}'
  ```
- [ ] **T22.2** - Validation with missing required fields
- [ ] **T22.3** - Validation with invalid data types

**Expected Results:**
- Validation correctly identifies schema compliance
- Clear error messages for validation failures
- Successful validation provides confidence for object creation

---

### Test 23: `workflows` - Explore available workflows and hint system

**Test Cases:**
- [ ] **T23.1** - List all workflows
  ```bash
  assets workflows --help
  assets workflows list
  ```
- [ ] **T23.2** - Show specific workflow details
  ```bash
  assets workflows show --workflow object_type_creation
  ```
- [ ] **T23.3** - Simulate workflow context
  ```bash
  assets workflows simulate --context create_object_type --variables '{"success":true}'
  ```

**Expected Results:**
- Workflow catalog shows all available workflows
- Workflow details provide step-by-step guidance
- Context simulation generates appropriate hints

---

## Contextual Hints Validation

### Global Hint Validation
**Test across all commands:**
- [ ] **H01** - Success scenarios generate appropriate next-step hints
- [ ] **H02** - Error scenarios provide helpful recovery hints  
- [ ] **H03** - Hints reference correct command syntax and parameters
- [ ] **H04** - Hints match current context and available data
- [ ] **H05** - Hint categories (essential, enhancement, maintenance) prioritized correctly

### Workflow Hint Chains
- [ ] **H06** - Object type creation → attribute addition → instance creation workflow
- [ ] **H07** - Attribute marketplace → extract → apply → validate workflow
- [ ] **H08** - Discovery → browse → trace → resolve workflow
- [ ] **H09** - Testing → create schema → populate → cleanup workflow

---

## Performance and Reliability Tests

### Performance Benchmarks
- [ ] **P01** - Schema listing under 2 seconds
- [ ] **P02** - Object listing (50 items) under 3 seconds
- [ ] **P03** - Attribute catalog (all schemas) under 10 seconds
- [ ] **P04** - Search operations under 5 seconds

### Error Handling
- [ ] **E01** - Network connectivity issues
- [ ] **E02** - Invalid authentication
- [ ] **E03** - Rate limiting responses
- [ ] **E04** - Invalid input data
- [ ] **E05** - Non-existent resource access

### Data Integrity
- [ ] **D01** - Object creation populates all required fields
- [ ] **D02** - Reference attributes maintain referential integrity
- [ ] **D03** - Schema operations don't affect other schemas
- [ ] **D04** - Test cleanup removes only test data

---

## Test Execution Plan

### Phase 1: Infrastructure Setup (T19)
1. Create test schema with sample data
2. Verify test environment isolation
3. Validate test data structure

### Phase 2: Core Operations (T01-T12, T14, T17, T21)
1. CRUD operations on test data
2. Search and filtering validation
3. Attribute management

### Phase 3: Advanced Features (T03-T05, T15-T16, T18, T20, T22-T23)
1. Intelligent completion and workflows
2. Schema management and tracing
3. Resolution and validation

### Phase 4: System Integration (T06-T07, T13)
1. Configuration management
2. Help system validation
3. Shell integration

### Phase 5: Contextual Hints (H01-H09)
1. Hint accuracy validation
2. Workflow chain testing
3. Context-appropriate suggestions

### Phase 6: Performance & Cleanup (P01-P04, E01-E05, D01-D04)
1. Performance benchmarking
2. Error scenario testing  
3. Test environment cleanup

---

## Success Criteria

### Functional Requirements
- [ ] All 23 command groups function correctly
- [ ] All contextual hints provide accurate guidance
- [ ] Error handling is helpful and actionable
- [ ] Performance meets acceptable thresholds

### Quality Requirements  
- [ ] Test coverage of all documented features
- [ ] Consistent behavior across commands
- [ ] Reliable operation in live environment
- [ ] Clean test data management

### Documentation Requirements
- [ ] Test results documented with pass/fail status
- [ ] Performance benchmarks recorded
- [ ] Known issues and limitations identified
- [ ] Recommendations for improvements documented

---

*Generated: $(date +%Y-%m-%d)  
Environment: Atlassian Assets CLI v1.0  
Workspace: d683300e-ec06-45ee-8789-b0f5e219c16f*