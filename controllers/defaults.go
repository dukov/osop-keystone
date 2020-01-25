/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package controllers

import (
	osconf "github.com/dukov/osop-common/pkg/openstack/config"
)

// KeystoneConfigDefaults default values for keystone.conf
var KeystoneConfigDefaults = osconf.IniFile{
	"DEFAULT": map[string]string{
		"max_token_size": "255",
		"transport_url":  "rabbit://user:password@rabbit",
	},
	"cache": map[string]string{
		"backend":         "dogpile.cache.memcached",
		"enabled":         "true",
		"memcach_servers": "memcached.default.svc.cluster.local:11211",
	},
	"credential": map[string]string{
		"key_repository": "/etc/keystone/credential-keys/",
	},
	"database": map[string]string{
		"connection":  "mysql+pymysql://keystone:password@mariadb.default.svc.cluster.local:3306/keystone",
		"max_retries": "-1",
	},
	"fernet_tokens": map[string]string{
		"key_repository": "/etc/keystone/fernet-keys/",
	},
	"identity": map[string]string{
		"domain_config_dir":               "/etc/keystonedomains",
		"domain_specific_drivers_enabled": "true",
	},
	"oslo_messaging_notifications": map[string]string{
		"driver": "messagingv2",
	},
	"oslo_messaging_rabbit": map[string]string{
		"rabbit_ha_queues": "false",
	},
	"oslo_middleware": map[string]string{
		"enable_proxy_headers_parsing": "true",
	},
	"security_compliance": map[string]string{
		"lockout_duration":         "1800",
		"lockout_failure_attempts": "5",
	},
	"token": map[string]string{
		"expiration": "43200",
		"provider":   "fernet",
	},
}
