# Changelog

All notable changes to this project will be documented in this file.
Versions are based on [Semantic Versioning](http://semver.org/), and the changelog is generated with [Chglog](https://github.com/git-chglog/git-chglog).

## Version History

* [v0.8.1](#v0.8.1)
* [v0.8.0](#v0.8.0)
* [v0.7.1](#v0.7.1)
* [v0.7.0](#v0.7.0)
* [v0.6.0](#v0.6.0)
* [v0.5.1](#v0.5.1)
* [v0.5.0](#v0.5.0)
* [v0.4.3](#v0.4.3)
* [v0.4.2](#v0.4.2)
* [v0.4.1](#v0.4.1)
* [v0.4.0](#v0.4.0)
* [v0.3.1](#v0.3.1)
* [v0.3.0](#v0.3.0)
* [v0.2.3](#v0.2.3)
* [v0.2.2](#v0.2.2)
* [v0.2.1](#v0.2.1)
* [v0.2.0](#v0.2.0)
* [v0.1.0](#v0.1.0)

## Changes

<a name="v0.8.1"></a>
### [v0.8.1](https://github.com/odit-services/s3ops/compare/v0.8.0...v0.8.1)

> 2025-08-15

#### 🏡 Chore

* update changelog
* **deploy:** update deployment manifests

#### 🚀 Enhancements

* **controller:** add error logging for Minio client health check

#### 💅 Refactors

* **api:** Support api tokens for servers as optional


<a name="v0.8.0"></a>
### [v0.8.0](https://github.com/odit-services/s3ops/compare/v0.7.1...v0.8.0)

> 2025-08-15

#### 🏡 Chore

* update changelog
* **deploy:** update deployment manifests

#### 🚀 Enhancements

* **api:** add ProviderMeta field to S3UserStatus for provider-specific metadata
* **client:** implement ListUsers method in IonosAdminClient for user retrieval
* **mocks:** enhance user existence checks with logging and detailed error messages

#### 🩹 Fixes

* **controller:** update user identifier condition for S3AdminClient type check

#### 💅 Refactors

* Add seperate client identifier for wider access key creation support
* **client:** simplify IsOnline method in IonosClient


<a name="v0.7.1"></a>
### [v0.7.1](https://github.com/odit-services/s3ops/compare/v0.7.0...v0.7.1)

> 2025-08-15

#### 🏡 Chore

* update changelog
* **deploy:** update deployment manifests

#### 💅 Refactors

* **client:** streamline client initialization using GenerateIonosClient


<a name="v0.7.0"></a>
### [v0.7.0](https://github.com/odit-services/s3ops/compare/v0.6.0...v0.7.0)

> 2025-08-15

#### 🏡 Chore

* update changelog
* **deploy:** update deployment manifests
* **deps:** Added dependency

#### 🚀 Enhancements

* **api:** Added baseline support for ionos
* **services:** Added usage of ionos s3 clients


<a name="v0.6.0"></a>
### [v0.6.0](https://github.com/odit-services/s3ops/compare/v0.5.1...v0.6.0)

> 2024-10-24

#### 🏡 Chore

* update changelog
* **deploy:** update deployment manifests

#### 🚀 Enhancements

* **api:** Add yaml to status
* **api:** Added new Name field to s3policy
* **s3policy_controller:** Implement policy name generation

#### 💅 Refactors

* **s3user_controller:** Switch to policy name generation

#### ✅ Tests

* **s3policy_controller:** Updated tests to match new naming rules
* **s3policy_controller:** Test for new policy status name field


<a name="v0.5.1"></a>
### [v0.5.1](https://github.com/odit-services/s3ops/compare/v0.5.0...v0.5.1)

> 2024-09-10

#### 🏡 Chore

* update changelog
* **deploy:** update deployment manifests

#### 🚀 Enhancements

* **s3bucket_controller:** Also create secret if it does not already exist


<a name="v0.5.0"></a>
### [v0.5.0](https://github.com/odit-services/s3ops/compare/v0.4.3...v0.5.0)

> 2024-09-10

#### 🏡 Chore

* update changelog
* **deploy:** update deployment manifests

#### 🚀 Enhancements

* **s3bucket_controller:** Implemented secret deletion
* **s3bucket_controller:** Implement secret creation

#### ✅ Tests

* **s3bucket_controller:** Test for correct type
* **s3bucket_controller:** Test secret deletion
* **s3bucket_controller:** Test for bucket connection details secret creation


<a name="v0.4.3"></a>
### [v0.4.3](https://github.com/odit-services/s3ops/compare/v0.4.2...v0.4.3)

> 2024-08-17

#### 🏡 Chore

* update changelog
* **deploy:** update deployment manifests

#### 🩹 Fixes

* **manager:** Increase memeory limit


<a name="v0.4.2"></a>
### [v0.4.2](https://github.com/odit-services/s3ops/compare/v0.4.1...v0.4.2)

> 2024-08-17

#### 🏡 Chore

* update changelog
* **deploy:** update deployment manifests

#### 🩹 Fixes

* **deployment:** Remove probles

#### 💅 Refactors

* **controller:** Remove unused Condition return values
* **controller:** Remove Condition return from helper functions
* **s3client:** Removed condition from helper functions


<a name="v0.4.1"></a>
### [v0.4.1](https://github.com/odit-services/s3ops/compare/v0.4.0...v0.4.1)

> 2024-08-17

#### 🏡 Chore

* update changelog
* **deploy:** update deployment manifests

#### 🩹 Fixes

* **controller:** Typo in policy


<a name="v0.4.0"></a>
### [v0.4.0](https://github.com/odit-services/s3ops/compare/v0.3.1...v0.4.0)

> 2024-08-17

#### 🏡 Chore

* update changelog
* **deploy:** update deployment manifests

#### 🚀 Enhancements

* **S3bucket_controller:** Implement user creation
* **api:** Added createuserfromtemplate for automagic user+policy generation
* **s3bucket_controller:** Added policy template for rw

#### 🩹 Fixes

* **api:** Allow empty as enum value

#### ✅ Tests

* **s3bucket_controller:** Test deletion of sample users
* **s3bucket_controller:** Test for call of secondary functions on user create
* **s3bucket_controller:** Test for correct policy
* **s3bucket_controller:** Test for user+policy creation


<a name="v0.3.1"></a>
### [v0.3.1](https://github.com/odit-services/s3ops/compare/v0.3.0...v0.3.1)

> 2024-08-17

#### 🏡 Chore

* update changelog
* **deploy:** update deployment manifests

#### 📖 Documentation

* Added some badges to the readme

#### 🚀 Enhancements

* **api:** Set printcolumns
* **controller:** Update lastaction according to the current state

#### 🩹 Fixes

* **rbac:** RBAC permissions for secrets


<a name="v0.3.0"></a>
### [v0.3.0](https://github.com/odit-services/s3ops/compare/v0.2.3...v0.3.0)

> 2024-08-17

#### 🏡 Chore

* update changelog
* **deploy:** update deployment manifests

#### 🚀 Enhancements

* **s3client:** Support external secrets for servers
* **s3server_controller:** Support for existing secrets

#### 💅 Refactors

* **api:** Add existingSecret ref option to s3server

#### ✅ Tests

* **s3bucket_controller:** Test for s3server with secretauth to handle s3client helper function test
* **s3server_controller:** Added tests for servers with existing secrets


<a name="v0.2.3"></a>
### [v0.2.3](https://github.com/odit-services/s3ops/compare/v0.2.2...v0.2.3)

> 2024-08-17

#### 🏡 Chore

* update changelog
* **deploy:** update deployment manifests
* **deploy:** update deployment manifests

#### 🚀 Enhancements

* **make:** Push for new releases


<a name="v0.2.2"></a>
### [v0.2.2](https://github.com/odit-services/s3ops/compare/v0.2.1...v0.2.2)

> 2024-08-17

#### 🏡 Chore

* update changelog
* **deploy:** update deployment manifests
* **deploy:** update deployment manifests

#### 🚀 Enhancements

* **make:** Patch version for yaml build


<a name="v0.2.1"></a>
### [v0.2.1](https://github.com/odit-services/s3ops/compare/v0.2.0...v0.2.1)

> 2024-08-17

#### 🏡 Chore

* update changelog
* update changelog
* update changelog
* update changelog
* update changelog

#### 📖 Documentation

* Added specific version deploy to readme
* Added missing information to readme

#### 🚀 Enhancements

* **changelog:** Updated template to include version index
* **deployment:** New build bundles
* **deployment:** Added full deployment build
* **make:** Added yaml bundle build target

#### 🩹 Fixes

* **changelog:** Fixed empty lines

#### 💅 Refactors

* **deploymeng:** Set default version and image


<a name="v0.2.0"></a>
### [v0.2.0](https://github.com/odit-services/s3ops/compare/v0.1.0...v0.2.0)

> 2024-08-17

#### 🏡 Chore

* **api:** New autogenerated functions
* **changelog:** Generated Changelog

#### 🚀 Enhancements

* **api:** Added CurrentRetries to CrStatus
* **api:** Added action consts
* **controller:** Added requeue on error
* **docs:** Updated changelog template
* **make:** Added release targtet
* **make:** Added advanced changelog generation script
* **s3bucket_controller:** Set new status fields
* **s3bucket_controller:** Info log for reconcile
* **samples:** Added user and policy samples
* **samples:** Added s3buckedt sample

#### 🩹 Fixes

* **api:** Make tls optional with default
* **api:** Make region optional
* **controller:** Don't reset retries
* **controller:** Add Event filters
* **s3bucket_controller:** Follow namingconventions
* **s3bucket_controller:** Don't override specialized status fields
* **s3bucket_controller:** Don't override default error
* **s3bucket_controller:** Fix truncate logic
* **s3bucket_controller:** Adjust nanoid alphabet to conform with minio standards
* **s3bucket_controller:** Truncate based on available length
* **s3user:** Cleanup secret at user deletion
* **s3user_controller:** Decode data from secret
* **samples:** Switch to default minio region
* **samples:** Switch to 100% localhost

#### 💅 Refactors

* **api:** Remove conditions from s3bucket status
* **api:** Switched from running to reconciling
* **api:** Remove conditions from S3User Status
* **api:** Add shared CrStatus
* **api:** Remove conditions from s3policy status
* **api:** Rename status.status to status.state
* **api:** Removed conditions from s3server
* **api:** Switched from Transition to Reconcile Time
* **controller:** Move ServerRef to shared object
* **s3bucket_controller:** Switch to the new state style ref [#1](https://github.com/odit-services/s3ops/issues/1)
* **s3bucket_controller:** Extract Status updates to HandleError function
* **s3policy_controller:** Switch to the new CrStatus from just conditions
* **s3server_controller:** Switch to CrStatus instead of conditions
* **s3user_controller:** Switch to writing CrStatus instead of conditions
* **s3user_controller:** Switch to nanoid user only

#### ✅ Tests

* **s3bucket_controller:** Fix typo in then
* **s3bucket_controller:** Test for the new CRStatus fields
* **s3bucket_controller:** Switch from state to condition
* **s3server_controller:** Test for CrStatus fields instead of condition magic
* **s3user_controller:** Switch to crstatus
* **s3user_controller:** Use correct secret data access syntax


<a name="v0.1.0"></a>
### v0.1.0

> 2024-08-16

#### 🏡 Chore

* **api:** new manifests

#### 📖 Documentation

* **README:** Added description and other facts

#### 🚀 Enhancements

* **api:** Added bucket spec info needed for creation
* **api:** New S3User Object
* **api:** Add name generation to bucket CR
* **api:** Added softdelete flag for s3bucket
* **api:** Basic S3Server Type
* **api:** Mark policy list as optional
* **api:** Added new cr bucket
* **api:** Added policy refs to user
* **api:** Added basics for s3policy
* **api:** Added s3bucket server reference
* **api:** Update required references for s3user
* **make:** Changelog generation
* **make:** Added multiarch target
* **s3bucket_controller:** Baseline setup
* **s3bucket_controller:** Added basic controller for s3 buckets
* **s3bucket_controller:** Implement bucket creation
* **s3bucket_controller:** Set correct finalizer
* **s3bucket_controller:** Implemented deletion
* **s3bucket_controller:** Fill in created field of the status
* **s3bucket_controller:** Implement name generation
* **s3client:** Added needed functions to interface
* **s3client:** Throw error when creating client from server config if server is offline
* **s3client:** Implemented policy apply
* **s3client:** Added minio implementation of s3client
* **s3client:** Baseline implementation for admin client
* **s3client:** Add interface functions for bucket operations
* **s3client:** Add RemoveBucket to interface
* **s3policy_controller:** Implemented policy deletion
* **s3policy_controller:** Implement s3policy creation
* **s3policy_controller:** Baseline Implementation
* **s3policy_controller:** implement outside actions for policy reconcile
* **s3server_controller:** Update online field
* **s3server_controller:** Implement s3server status check on ressource creation
* **s3server_controller:** Requeue
* **s3user_controller:** Implemented secret generation
* **s3user_controller:** Added additional logs/status conditions
* **s3user_controller:** Basics of secret handling
* **s3user_controller:** Base setup
* **s3user_controller:** Implement user creation
* **s3user_controller:** Implemented deletion
* **s3user_controller:** Implement policy assign

#### 🩹 Fixes

* **api:** Moved object annotation
* **docker:** Copy over services directory
* **make:** Typo in multiarch target
* **mocks:** pass on spy
* **s3bucket_controller:** Always return errors if thrown
* **s3policy_controller:** Typo
* **s3server_controller:** Start healthcheck before cheking online status
* **s3server_controller:** Added Reason and lastupdate to all status updates
* **s3user_controller:** Confirm with naming standards
* **s3user_controller:** Generate password with correct settings)



* **api:** Switch to disable flag instead of enable flag for s3bucket name generation

#### 💅 Refactors

* **api:** Switch to seperate status functions
* **api:** Remove ability to provide s3user custom secret
* **api:** Remove Port
* **api:** Use sharted cr status instead of implementing status for all CRs
* **controller:** Replace default minio client with mockable interface
* **controller:** Extract s3client related stuff into it's own module
* **controllers:** Extract s3 server get into shared function
* **controllers:** Rename to reflect job instead of provider
* **s3client:** Handle different client implementations
* **s3client:** Move to services folder
* **s3server_controller:** Extract minio client generation to shared function
* **s3user_controller:** Removed logic related to user-provided secrets
* **s3user_controller:** Migrate secret handling to helper functions
* **tests:** Use mocked s3client instead of guesswork with play.min.io

#### 🎨 Styles

* **s3bucket_controller:** Removed unused comments
* **s3server_controller:** Typo

#### ✅ Tests

* **controllers:** Added spy to s3client mocks
* **controllers:** Reconcile test server objects for non-server reconcilers
* **controllers:** Fix a bunch of nill pointer exceptions
* **mocks:** Mock S3Client
* **s3bucket:** Test for finalizer being set
* **s3bucket_controller:** Check name generation
* **s3bucket_controller:** Added tests for happy path
* **s3bucket_controller:** Added deletion tests
* **s3bucket_controller:** reset spy and mark bucket as existing
* **s3bucket_controller:** Typo in object ref
* **s3bucket_controller:** Added tests for s3 server problems
* **s3bucket_controller:** Envoke deletion for delete tests
* **s3bucket_controller:** Check new created status field
* **s3client:** Implemente empty mock bodies
* **s3client:** Added admin mocks
* **s3client:** Implement mocks
* **s3policy_controller:** Test policy deletion
* **s3policy_controller:** Basic test setup
* **s3policy_controller:** Test policy updates
* **s3server_controller:** Implemented basic tests
* **s3server_controller:** Check condition for all tests
* **s3server_controller:** Test updates
* **s3server_controller:** Test deletion
* **s3server_controller:** Test for new online status
* **s3user_controller:** Added deletion happy path test
* **s3user_controller:** Create test policy for user
* **s3user_controller:** Test for policy apply
* **s3user_controller:** Also enforce checkpolicy
* **s3user_controller:** Switch to loading policy name from env
* **s3user_controller:** Fix in bracket order
* **s3user_controller:** Added tests for invalid servers
* **s3user_controller:** Test basic controller functions
* **s3user_controller:** Test happy path for create and update

