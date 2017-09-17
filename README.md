check-rabbitmq
==============

Sensu / Nagios compatible plugin to check RabbitMQ queue item counts. 
Requires RabbitMQ management plugin to be enabled:

	$ rabbitmq-plugins enable rabbitmq_management


Usage
=====

	$ check-rabbitmq -host http://somehost:15672 \
		-user admin -password secret 
		-error 10 -warn 5 \
		-queue 'pattern' \
		-exclude '^some_queue$'

Why
===
I needed an excuse to learn golang.
