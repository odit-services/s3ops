
<a name="v0.1.0"></a>
## v0.1.0

> 2024-08-16

### Chore

* **api:** new manifests

### Docs

* **README:** Added description and other facts

### Feat

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

### Fix

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

### Refacotor

* **api:** Switch to disable flag instead of enable flag for s3bucket name generation

### Refactor

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

### Style

* **s3bucket_controller:** Removed unused comments
* **s3server_controller:** Typo

### Test

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

