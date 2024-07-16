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
	"S3 IA",
	"S3",
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

// ToAttributesCase Converts any interface to SDP attributes and also fixes case
// to be the correct `camelCase`. Excluded fields can also be provided, the
// field names should be provided in the final camelCase format. Arrays are also
// sorted to ensure consistency.
func ToAttributesCase(i interface{}, exclusions ...string) (*sdp.ItemAttributes, error) {
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

	camel := CamelCase(m)

	if camelMap, ok := camel.(map[string]interface{}); ok {
		for _, exclusion := range exclusions {
			// Exclude some things
			delete(camelMap, exclusion)
		}
		return sdp.ToAttributesSorted(camelMap)
	} else {
		return &sdp.ItemAttributes{}, errors.New("could not convert camel cased data to map[string]interface{}")
	}
}

// CamelCase converts all keys in an object to camel case recursively, this
// includes ignoring known acronyms
func CamelCase(i interface{}) interface{} {
	if i == nil {
		return nil
	}

	v := reflect.ValueOf(i)

	switch v.Kind() {
	case reflect.Map:
		newMap := make(map[string]interface{})

		iter := v.MapRange()
		for iter.Next() {
			k := iter.Key()
			v := iter.Value()
			vi := v.Interface()

			if vi != nil {
				keyCamel := strcase.ToLowerCamel(k.String())

				newMap[keyCamel] = CamelCase(vi)
			}
		}

		return newMap
	case reflect.Array, reflect.Slice:
		newSlice := make([]interface{}, 0, v.Len())

		for index := 0; index < v.Len(); index++ {
			newSlice = append(newSlice, CamelCase(v.Index(index).Interface()))
		}

		return newSlice
	default:
		return i
	}
}
