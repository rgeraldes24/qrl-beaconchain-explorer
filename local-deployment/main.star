qrl_module = import_module("github.com/theQRL/qrl-package/main.star")

POSTGRES_PORT_ID = "postgres"
POSTGRES_DB = "db"
POSTGRES_USER = "postgres"
POSTGRES_PASSWORD = "pass"

POSTGRESADM_PORT_ID = "postgresadm"

REDIS_PORT_ID = "redis"

REDISINSIGHT_PORT_ID = "redisinsight"

LITTLE_BIGTABLE_PORT_ID = "littlebigtable"

def run(plan, args):
	db_services = plan.add_services(
		configs={
			# Add a Postgres server
			"postgres": ServiceConfig(
				image = "postgres:15.2-alpine",
				ports = {
					POSTGRES_PORT_ID: PortSpec(5432, application_protocol = "postgresql"),
				},
				env_vars = {
					"POSTGRES_DB": POSTGRES_DB,
					"POSTGRES_USER": POSTGRES_USER,
					"POSTGRES_PASSWORD": POSTGRES_PASSWORD,
				},
			),
			# Add Postgres Admin
			"postgresadm": ServiceConfig(
			 	image = "dpage/pgadmin4",
				ports = {
					POSTGRESADM_PORT_ID: PortSpec(80, application_protocol = "tcp"),
				},
				env_vars = {
					"PGADMIN_DEFAULT_EMAIL": "test@test.com",
					"PGADMIN_DEFAULT_PASSWORD": "pass",
				},
			),
			# Add a Redis server
			"redis": ServiceConfig(
				image = "redis:7",
				ports = {
					REDIS_PORT_ID: PortSpec(6379, application_protocol = "tcp"),
				},
			),
			# Add RedisInsight
			"redisinsight": ServiceConfig(
				image = "redis/redisinsight:latest",
				ports = {
					REDISINSIGHT_PORT_ID: PortSpec(5540, application_protocol = "tcp"),
				},
			),
			# Add a Bigtable Emulator server
			"littlebigtable": ServiceConfig(
				image = "gobitfly/little_bigtable:latest",
				ports = {
					LITTLE_BIGTABLE_PORT_ID: PortSpec(9000, application_protocol = "tcp"),
				},
			),
		}
	)
	
	# Spin up a local qrl testnet
	qrl_module.run(plan, args)