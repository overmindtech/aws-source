package sources

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/iancoleman/strcase"
	"github.com/overmindtech/sdp-go"
)

var Acronyms = []string{
	"ACL",
	"ADFS",
	"AES",
	"AI",
	"AMI",
	"API",
	"ARN",
	"ASG",
	"AWS",
	"AZ",
	"CDN",
	"CIDR",
	"CLI",
	"CORS",
	"DaaS",
	"DDoS",
	"DMS",
	"DNS",
	"DoS",
	"EBS",
	"EC2",
	"ECS",
	"EFS",
	"EIP",
	"ELB",
	"EMR",
	"ENI",
	"FaaS",
	"FIFO",
	"HPC",
	"HTTP",
	"HTTPS",
	"HVM",
	"IaaS",
	"IAM",
	"ICMP",
	"IGW",
	"IOPS",
	"IOT",
	"IP",
	"IPSec",
	"iSCSI",
	"JSON",
	"KMS",
	"LB",
	"MFA",
	"MITM",
	"MPLS",
	"MPP",
	"MSTSC",
	"NAT",
	"NFS",
	"NS",
	"OLAP",
	"OLTP",
	"PaaS",
	"PCI DSS",
	"PV",
	"RAID",
	"RAM",
	"RDS",
	"RRS",
	"S3",
	"S3 IA",
	"SaaS",
	"SaaS",
	"SAML",
	"SDK",
	"SES",
	"SLA",
	"SMS",
	"SNS",
	"SOA",
	"SOAP",
	"SQS",
	"SSE",
	"SSH",
	"SSL",
	"SSO",
	"STS",
	"SWF",
	"TCP",
	"TLS",
	"TPM",
	"TPM",
	"TPS",
	"TTL",
	"VDI",
	"VLAN",
	"VM",
	"VPC",
	"VPG",
	"VPN",
	"VTL",
	"WAF",
}

func init() {
	for _, acronym := range Acronyms {
		// Load acronyms so tha they won't be wrecked by camelCase
		strcase.ConfigureAcronym(acronym, acronym)
	}
}

// ToAttributesCase Converts any interace to SDP attributes and also fixes case
// to be the correct `camelCase`
func ToAttributesCase(i interface{}) (*sdp.ItemAttributes, error) {
	var m map[string]interface{}

	// Convert via JSON
	b, err := json.Marshal(i)

	if err != nil {
		return &sdp.ItemAttributes{}, err
	}

	err = json.Unmarshal(b, &m)

	if err != nil {
		return &sdp.ItemAttributes{}, err
	}

	camel := CamelCaseMap(m)

	if camelMap, ok := camel.(map[string]interface{}); ok {
		return sdp.ToAttributes(camelMap)
	} else {
		return &sdp.ItemAttributes{}, errors.New("could not convert camel cased data to map[string]interface{}")
	}
}

// CamelCaseMap converts all keys in a map to camel case recursively, this
// includes ignoring known acronyms
func CamelCaseMap(m interface{}) interface{} {
	if m == nil {
		return nil
	}

	newMap := make(map[string]interface{})

	v := reflect.ValueOf(m)
	t := reflect.TypeOf(m)

	// If it's not a mep then we can't do anything
	if t.Kind() != reflect.Map {
		return m
	}

	iter := v.MapRange()
	for iter.Next() {
		k := iter.Key()
		v := iter.Value()
		vi := v.Interface()

		if vi != nil {
			keyCamel := strcase.ToLowerCamel(k.String())

			newMap[keyCamel] = CamelCaseMap(vi)
		}
	}

	return newMap
}
