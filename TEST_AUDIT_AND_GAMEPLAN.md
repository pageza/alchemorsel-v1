# Test Audit and Game Plan for MVP Standardization

## I. Overview

**Objective:** Achieve a consistent test environment for our MVP by aligning test expectations across unit, integration, and end-to-end (e2e) tests. *Peripheral tests (chaos, load, mutation, performance, security) will be temporarily skipped (commented out) until core functionality is stabilized.*

## II. Core Issues Identified

### 1. Data Model Discrepancies:
- **Issue:** The recipe object is inconsistent. Some unit tests expect the `steps` field as a string, while integration/e2e tests expect it as an array or structured object.
- **Recommendation:** Decide on a single structure (e.g., use an array) and update both the implementation and all tests.

### 2. Authentication & Authorization:
- **Issue:** Unit tests bypass authentication (using mocks or context pre-setting) whereas integration and e2e tests enforce JWT token validation and proper error codes (e.g., 401 Unauthorized).
- **Recommendation:** Standardize on a unified authentication contract—either simulate a valid JWT token in all tests or enforce proper token validation with a common test bypass when needed.

### 3. Error Handling & Middleware:
- **Issue:** Discrepancy in error responses. Some tests expect plain error messages, others expect a JSON object with `code` and `message`.
- **Recommendation:** Define a single error response format (JSON with `code` and `message`) and update all tests accordingly.

### 4. Rate Limiting Configuration:
- **Issue:** Inconsistent rate limiter behavior across tests. Test mode configuration (TestConfig) must be uniformly applied.
- **Recommendation:** Ensure that all tests run with the same environment variables (e.g., `INTEGRATION_TEST=true`) so that rate limits are relaxed consistently.

### 5. Mocks and Dependency Injection:
- **Issue:** Outdated mock signatures in unit tests (argument order mismatches, missing context parameters).
- **Recommendation:** Update mock setups to mirror the actual dependencies and handler implementations.

## III. Game Plan for MVP

1. **Standardize API Contract:**
   - Define and document the unified data model for key objects (e.g., Recipe, User).
   - Set a standard error response format and authentication process for all endpoints.

2. **Update Core Business Logic and Tests:**
   - Align the recipe model: decide on the `steps` field format and update both implementation and tests.
   - Refactor authentication in unit and integration tests to use a consistent token generation or bypass mechanism.
   - Update error handling in all test scenarios to expect consistent JSON responses.
   - Adjust mock setups in unit tests to match the actual function signatures and dependency behaviors.

3. **Disable Peripheral Tests Temporarily:**
   - Identify peripheral tests (chaos, load, mutation, performance, security) and add `t.Skip()` calls or comment them out.
   - Document the specific files/directories to skip so they can be re-enabled later.

4. **Unified Test Configuration:**
   - Establish environment variables (e.g., `INTEGRATION_TEST`, `TEST_RATE_LIMIT_STRICT`) for use in tests to apply the TestConfig uniformly.

## IV. Detailed Audit of Test Files

### A. Core Tests:
1. **internal/handlers/recipe_handler_test.go**
   - *Discrepancy:* Expects recipe.steps as a string; implementation sends an array/structured data.
   - *Action:* Update either the model or test expectations for consistency.

2. **internal/handlers/user_handler_test.go**
   - *Discrepancy:* Authentication bypass used in unit tests vs. enforced JWT validation in integration tests.
   - *Action:* Standardize authentication handling (e.g., by injecting a test user or using a common test token).

3. **tests/e2e/collections/recipes/recipe_endpoints.postman_collection.json**
   - *Discrepancy:* Requires a properly formatted JWT token and expects detailed recipe fields.
   - *Action:* Ensure that the API response format matches the documented contract in the e2e tests.

4. **tests/e2e/collections/users/user_endpoints.postman_collection.json**
   - *Discrepancy:* Expectations for additional fields (e.g., bio, preferences) that are not part of the core model.
   - *Action:* Reconcile the user model documented in the core business logic with expectations in e2e tests.

### B. Peripheral Tests (to be skipped for now):
1. **Chaos Tests (e.g., tests/chaos/*)**
   - *Action:* Comment out test bodies or add `t.Skip()` statements to temporarily disable these tests.

2. **Load Tests (e.g., tests/load/*)**
   - *Action:* Similarly, mark these as skipped for MVP.

3. **Mutation, Performance, and Security Tests**
   - *Action:* Disable with `t.Skip()` while core functionality is prioritized.

## V. Next Steps

1. Use this document as a central reference when making changes to ensure consistency across tests.
2. Update core tests and implementation in tandem to achieve a unified set of expectations.
3. Once core logic is stable and tests pass, gradually re-enable peripheral tests and align their expectations as needed. 



