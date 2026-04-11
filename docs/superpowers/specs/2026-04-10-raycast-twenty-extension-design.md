# Raycast Twenty Extension Rewrite Design

Date: 2026-04-10
Status: Approved in chat, written for review
Owner: Codex session in `twenty-cli`

## Summary

Rework the upstream Raycast `extensions/twenty` extension through a sequence of small pull requests.

The target outcome is:

- keep `extensions/twenty` as the upstream extension instead of creating a forked store entry
- add correct self-hosted Twenty support through a base URL preference
- internally replace the current narrow `TwentySDK` implementation with a typed client/service layer
- preserve and improve generic native Twenty object-record creation
- add a Google-Contacts-style people workflow backed by the native Twenty `people` object

This project copies the successful UX patterns from `extensions/google-contacts`, but the source of truth remains the standard Twenty schema, especially the native `people` object.

## Decision

Implement this as a sequence of focused upstream PRs against `raycast/extensions`, not as:

- a standalone forked extension
- a one-shot large PR
- a thin patch on top of the current `extensions/twenty` internals

The current `extensions/twenty` should be treated as an upstream target with partially reusable UI surface, but not as an architecture to preserve.

## Context And Research

### Upstream `extensions/twenty`

Current state:

- active upstream extension with recent maintenance
- only supports a single `Create Object Record` command
- already exposes a token preference and a URL preference
- current implementation is built around a narrow `TwentySDK` wrapper and object-create form flow

Problems in the current implementation:

- service layer returns weak result types like `string | boolean`
- base URL handling is simplistic and not robust for hosted versus self-hosted normalization
- record creation flow hard-codes assumptions about primary fields such as `name` and `title`
- the extension architecture is not set up for browse/search/detail/edit workflows

### Upstream `extensions/google-contacts`

`google-contacts` provides the UX reference to port:

- dedicated search/browse command
- list/detail and grid browsing modes
- componentized actions and forms
- create/edit flows built around a consistent client and hook layer
- quick-add command

This extension is structurally closer to the desired Raycast behavior than the current `extensions/twenty`.

### Local `twenty-cli`

The local `twenty-cli` repo contains cleaner primitives than the current Raycast Twenty extension:

- workspace and base URL resolution
- reusable HTTP client with retries and debug support
- metadata services for objects and fields
- generic records CRUD services
- search service for Twenty GraphQL search

These patterns should be ported into the Raycast extension, but the Raycast extension should not directly depend on the external `@salmonumbrella/twenty-cli` package at runtime.

## Goals

- Upstream work into `raycast/extensions` through small PRs
- Support self-hosted Twenty instances via explicit base URL preference
- Make the extension work cleanly against the native Twenty `people` object
- Port the successful contact-management UX patterns from `google-contacts`
- Improve the generic native-object story of the existing Twenty extension where it is part of the promised scope

## Non-Goals

- creating a new standalone Raycast extension with a separate store identity
- inventing a parallel local sync layer outside Twenty
- forcing a Google People data model onto Twenty where the native schema differs
- blocking the people workflow on every possible generic-object enhancement

## Product Scope

The native Twenty `people` object is the contact source of truth.

The extension should present people/contact workflows that feel similar to `google-contacts`, but field mapping must stay faithful to Twenty. The extension should only expose behaviors that map cleanly to standard Twenty data.

## Architecture

### Extension Boundary

The Raycast extension should contain its own extension-native client layer.

It should:

- reuse design patterns from `twenty-cli`
- copy or adapt the relevant client/service logic into `extensions/twenty`
- avoid a runtime dependency on a package published from another repository

Reasons:

- upstream maintainers will expect the extension to be self-contained in the Raycast repo
- release/versioning would be awkward if the extension depended on an external package
- the Raycast extension needs a smaller, UI-oriented surface than the full CLI package

### Client Layer

PR 1 should introduce a typed internal service layer with responsibilities roughly split as:

- config and preference resolution
- URL normalization
- HTTP client and error translation
- metadata service for objects and fields
- records service for list/get/create/update/delete
- people-oriented service helpers where needed
- search service where useful for browse/find experiences

The client layer should normalize:

- hosted Twenty URLs
- self-hosted URLs
- input values with or without trailing slash
- input values that may already include `/rest`

The client layer should not silently fall back to production hosted Twenty when the user supplied an invalid self-hosted URL. Invalid user input should surface as configuration error.

### Generic Object Handling

The rewritten generic object-create flow should be metadata-driven instead of object-name-driven.

Where the existing extension has TODOs for native field support, PR 1 should implement the native field types that are straightforward to support in the generic object-create flow:

- phone number
- currency
- boolean
- rating
- multi-select

Relation support should be treated as a separate complexity tier and explicitly deferred to PR 3 unless the rewritten metadata and form layer make it trivial during implementation.

The implementation should prioritize field support that is native to Twenty and useful for both generic object creation and the `people` object.

### Contact UX Mapping

PR 2 should add a people workflow inspired by `google-contacts`.

Expected commands:

- `Search People`
- `Create Person`
- `Quick Add Person`

The main `Search People` command should support a browse/search/detail workflow that feels close to `google-contacts`, adapted to Twenty’s schema.

Expected behaviors:

- search and browse people records
- open detail view for a selected person
- edit person from the result list
- create a new person from a full form
- quick-add a person from arguments or a compact form
- action shortcuts for relevant contact operations when the fields exist

Field mapping should remain Twenty-native. Examples of likely mappings:

- display name: native person name fields
- email addresses: Twenty email fields if present
- phone numbers: Twenty phone fields if present
- company or title: native company/title-related fields if present
- links and socials: Twenty links fields
- notes: text field if present
- avatar: image/avatar field if available

Not every Google Contacts feature should be mirrored. If a feature does not map cleanly to native Twenty `people`, it should be omitted or adapted instead of faked.

## Delivery Plan

### PR 1: Foundation Rewrite And Self-Hosted Support

Scope:

- add and validate explicit base URL preference handling
- rewrite the current `TwentySDK` into a typed internal client/service layer
- clean up error handling and configuration failures
- rebuild generic object-create flow on top of the new services
- finish generic native field support for phone number, currency, boolean, rating, and multi-select

Exit criteria:

- hosted and self-hosted instances both work
- no object-name hacks for primary-field assumptions remain in the core create flow unless forced by missing metadata
- generic create flow remains functional

### PR 2: People Contact Workflow

Scope:

- port the interaction model from `google-contacts` into `extensions/twenty`
- add browse/search/detail views for `people`
- add create/edit person flow
- add quick-add person command
- add person actions that map cleanly to native Twenty data

Exit criteria:

- extension can be used as a contact browser/editor for a normal Twenty `people` dataset
- self-hosted instances work the same as hosted ones

### PR 3: Extension Expansion

Scope:

- add relation-field support for generic native objects if it did not land earlier
- finish remaining native generic-object support not required for PR 1
- consider broader record viewing across objects if it still fits the extension after the people workflow lands
- close out remaining native-field or object UX gaps

Exit criteria:

- old upstream TODOs that still matter are either implemented or intentionally removed from the roadmap

## Testing Strategy

Each PR should include targeted verification appropriate to its scope.

### PR 1

- preference parsing and URL normalization
- hosted URL and self-hosted URL request construction
- service-level tests for metadata and record flows
- manual smoke test for create-object flow

### PR 2

- people search and browse behavior
- person create and edit flows
- quick-add flow
- self-hosted smoke test against a real local Twenty instance

### PR 3

- native field rendering and submission tests for newly supported field types
- generic-object viewing or listing tests if added

## Risks

- Twenty metadata may still not expose enough information to infer every ideal form behavior cleanly
- relation-field UX in Raycast may require more design work than other field types
- the native `people` schema may not match every `google-contacts` affordance one-to-one
- upstream maintainers may prefer narrower PRs than expected, so each PR should stay independently useful

## Open Decisions Already Resolved

- Upstream target: `extensions/twenty`
- Contact source of truth: native standard Twenty `people` object
- Delivery style: sequence of smaller PRs
- Architecture: internal rewrite using `twenty-cli` patterns, not direct package dependency

## Next Step

After review of this design doc, create an implementation plan for PR 1 only.
