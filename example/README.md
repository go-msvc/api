# Users API

# Tables #
- accounts
- users
- sessions

System admin user has account_id=0 and admin=1
Account users has account_id > 0 and belongs to that account.
If account user is admin, it can create more account users.

DONE:
- GET /users uses token to limit view on own account only (admin see all)

TODO:
- apply claim in other api, e.g. get accounts and add-user may only add for own account and not require account id in request
- Add field to account.max_users (e.g. pay more to allow bigger company) (store in acc or generic profile table?)
- Reset password by admin to specified password or by user with password sent to email/sms
- extend must respond with more info, i.e. updated session info
- add roles/groups and permissions that users can belong to also system/per account
- upd/del users/accounts
- logout to end session
- proper logs/reports
- option to add salt to password hash - must also apply to admin password, from environment var
