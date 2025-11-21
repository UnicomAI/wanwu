## HTTP Header
| Header        | Description   |
| ------------- | ------------- |
| Authorization | JWT token |
| X-Language    | Language code |
| X-Org-Id      | Organization ID |
| X-Client-Id   | Client identifier |

## HTTP Status
| HTTP Status             | Description                 |
| ----------------------- | -------------------------- |
| 200, StatusOK           | Request returned successfully|
| 400, StatusBadRequest   | Request failed, for business |
| 401, StatusUnauthorized | JWT authentication failed |
| 403, StatusForbidden    | No permission             |

## Permission-Menu Mapping Table
| First Level Permission | Second Level Permission | Third Level Permission | First Level Menu | Second Level Menu | Third Level Menu |
|----------------------|----------------------|-----------------------|------------------|------------------|------------------|
| guest                 |                      |                       | 【Guest】         |                  |                  |
| common                |                      |                       | 【Common】        |                  |                  |
| permission            |                      |                       | Permission Management |                  |                  |
| permission            | user                 |                       | Permission Management | User Management |                  |
| permission            | org                  |                       | Permission Management | Organization Management |                  |
| permission            | role                 |                       | Permission Management | Role Management |                  |

## `/v1/user/permission` Response Example
```json
{
  "code": 0,
  "data": {
    "orgPermission": {
      "org": {"id": "test-org-id", "name": "test-org-name"},
      "permissions": [
        {"perm": "permission"},
        {"perm": "permission.user"},
        {"perm": "permission.org"},
        {"perm": "permission.role"}
      ]
    }
  },
  "msg": "Operation successful"
}
```