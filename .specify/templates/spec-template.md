# Feature Specification: [FEATURE NAME]

**Feature Branch**: `feature/[descriptive-name]`
**Created**: [DATE]
**Status**: Draft
**Input**: User description: "$ARGUMENTS"

## User Scenarios & Testing *(mandatory)*

<!--
  IMPORTANT: User stories should be PRIORITIZED as user journeys ordered by importance.
  Each story must be INDEPENDENTLY TESTABLE — if only one is implemented,
  you should still have a viable MVP that delivers value.

  Tech Stack Constraint: loopi-api uses stdlib `testing` with hand-written fakes.
  Acceptance scenarios must be expressible as Go unit/integration tests.
-->

### User Story 1 - [Brief Title] (Priority: P1)

[Describe this user journey in plain language]

**Why this priority**: [Explain the value and why it has this priority level]

**Independent Test**: [Describe how this can be tested via Go tests — e.g., "Service method `X` tested with a fake repo returning `Y`"]

**Acceptance Scenarios**:

1. **Given** [initial state], **When** [action on service/handler], **Then** [expected outcome / return value]
2. **Given** [initial state], **When** [action], **Then** [expected outcome]

---

### User Story 2 - [Brief Title] (Priority: P2)

[Describe this user journey in plain language]

**Why this priority**: [Explain the value and why it has this priority level]

**Independent Test**: [Describe how this can be tested independently]

**Acceptance Scenarios**:

1. **Given** [initial state], **When** [action], **Then** [expected outcome]

---

[Add more user stories as needed, each with an assigned priority]

### Edge Cases

- What happens when [boundary condition, e.g., resource not found → `apperrors.ErrNotFound`]?
- How does the system handle [error scenario, e.g., DB failure → `apperrors.ErrInternalServer`]?
- What happens when [concurrent / duplicate request]?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST [specific capability]
- **FR-002**: System MUST [specific capability]
- **FR-003**: Users MUST be able to [key interaction]

*Example of marking unclear requirements:*

- **FR-00X**: System MUST [NEEDS CLARIFICATION: unspecified detail]

### Key Entities *(include if feature involves data)*

- **[Entity]**: [What it represents, key attributes — no implementation details]

## Integration Points with Existing System

**This feature interacts with the following existing modules**:

| Module | Path | Interaction Type |
|--------|------|-----------------|
| [Existing Service] | `internal/application/service/<domain>_service.go` | [Calls / Called by] |
| [Existing Repository] | `internal/domain/repository/<domain>_repository.go` | [Interface to implement] |
| [Existing Handler] | `internal/interface/handler/<domain>/` | [Extends / New route] |

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: [Measurable metric, e.g., "All acceptance scenarios pass as Go tests"]
- **SC-002**: [Measurable metric, e.g., "New code coverage ≥ 80%"]
- **SC-003**: [Business metric]

## Assumptions

- [Assumption about scope, e.g., "Admin-only endpoint, protected by `middleware.AdminOnly`"]
- [Assumption about data, e.g., "Entity already exists in `internal/domain/entity/`"]
- [Dependency, e.g., "Reuses existing `MySQLXxxRepository` for data access"]
