db.createUser({
	user: "mongo",
	pwd: "password",
	roles: [
		{ role: "readWrite", db: "acmeair" },
		{ role: "dbAdmin", db: "acmeair" },
	]
});
