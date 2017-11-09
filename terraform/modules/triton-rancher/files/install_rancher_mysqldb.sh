#!/bin/bash

set -o xtrace

if [ -n "$(command -v yum)" ]; then
	# CentOS
	wget https://dev.mysql.com/get/mysql57-community-release-el7-9.noarch.rpm
	sudo rpm -ivh mysql57-community-release-el7-9.noarch.rpm
	rm mysql57-community-release-el7-9.noarch.rpm
	sudo yum update -y
	sudo yum install mysql-server -y

	sudo systemctl start mysqld
	mysql_temp_password=$(sudo grep 'temporary password' /var/log/mysqld.log | tail -1 | sed -n 's/.*root@localhost: //p')

	# Update mysql bind-address
	sudo sh -c 'printf "bind-address=0.0.0.0\nport=${mysqldb_port}\nvalidate-password=off\n" >> /etc/my.cnf'
	sudo systemctl restart mysqld

	# Set mysql root password
	mysql -uroot -p$mysql_temp_password --connect-expired-password -Bse "ALTER USER 'root'@'localhost' IDENTIFIED BY '${mysqldb_password}';"

	# Open firewalld
	firewall-cmd --zone=public --add-service=mysql --permanent
	firewall-cmd --reload

elif [ -n "$(command -v apt-get)" ]; then
	# Ubuntu
	export DEBIAN_FRONTEND="noninteractive"

	echo debconf mysql-server/root_password password ${mysqldb_password} | sudo debconf-set-selections
	echo debconf mysql-server/root_password_again password ${mysqldb_password} | sudo debconf-set-selections

	sudo apt-get update -y
	sudo apt-get -qq install mysql-server-5.7 -y

	sudo apt-get update -y
	sudo apt-get install expect -y

	MYSQL_ROOT_PASSWORD=${mysqldb_password}

	SECURE_MYSQL=$(expect -c "
	set timeout 10
	spawn mysql_secure_installation
	expect \"Enter current password for root:\"
	send \"$MYSQL_ROOT_PASSWORD\r\"
	expect \"Would you like to setup VALIDATE PASSWORD plugin?\"
	send \"n\r\" 
	expect \"Change the password for root ?\"
	send \"n\r\"
	expect \"Remove anonymous users?\"
	send \"y\r\"
	expect \"Disallow root login remotely?\"
	send \"y\r\"
	expect \"Remove test database and access to it?\"
	send \"y\r\"
	expect \"Reload privilege tables now?\"
	send \"y\r\"
	expect eof
	")

	echo "$SECURE_MYSQL"

	# Update mysql bind-address
	sudo sh -c 'printf "[mysqld]\nbind-address=0.0.0.0\nport=${mysqldb_port}\n" > /etc/mysql/my.cnf'
	sudo service mysql restart

else
	echo "Unsupported platform";
	exit 1;
fi

# Create db, user, and setup permissions
mysql -u root -p${mysqldb_password} --port ${mysqldb_port} <<MYSQL_SCRIPT
CREATE DATABASE IF NOT EXISTS ${mysqldb_database_name} COLLATE = 'utf8_general_ci' CHARACTER SET = 'utf8';
GRANT ALL ON ${mysqldb_database_name}.* TO '${mysqldb_user}'@'%' IDENTIFIED BY '${mysqldb_password}';
GRANT ALL ON ${mysqldb_database_name}.* TO '${mysqldb_user}'@'localhost' IDENTIFIED BY '${mysqldb_password}';
FLUSH PRIVILEGES;
MYSQL_SCRIPT
