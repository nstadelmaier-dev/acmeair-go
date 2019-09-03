db.createUser({
	user: "admin",
	pwd: "password",
	roles: [ { role: "userAdminAnyDatabase", db: "admin" } ]
});
