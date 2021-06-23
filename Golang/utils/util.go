package utils

import (
	"fmt"
	"strings"
)

var listValueCountryVN = map[string]bool{
	"vietnam":  true,
	"VietNam":  true,
	"vietNam":  true,
	"Vietnam":  true,
	"VIETNAM":  true,
	"việt nam": true,
	"Việt Nam": true,
	"VIỆT NAM": true,
	"VN":       true,
	"Vn":       true,
	"vn":       true,
	"vN":       true,
}

var listSeparated = []string{
	",", // Sample Address: XXX, YYY, ZZZ
	"-", // Sample Address: XXX - YYY - ZZZ
	"/", // Sample Address: XXX / YYY / ZZZ
}

type AddressDetail struct {
	FullAddress string
	Street      string
	Ward        string
	District    string
	City        string
	Country     string
}

func StandardizedAddress(fullAddress string) {

}

func splitFullAddress(fullAddress, separated string) []string {
	return strings.Split(fullAddress, separated)
}

func GetAddressDetailFromFullAddress(fullAddress string) (addressDetail AddressDetail) {
	addressDetail.FullAddress = fullAddress
	if len(fullAddress) == 0 {
		return
	}
	var arrFullAddress = []string{}
	for _, value := range listSeparated {
		arrFullAddress = splitFullAddress(fullAddress, value)
		if len(arrFullAddress) > 1 {
			break
		}
	}

	// Format names
	// 0,1,2: Format invalid
	// 4: Sample format: 7/28 Thanh Thai, Quận 10, TPHCM, VN || 7/28 Thanh Thai, Phường 4, Quận 10, TPHCM
	// 5: Sample format: 7/28 Thanh Thai, Phường 14, Quận 10, TPHCM, VN || Tòa nhà Rivera Park, 7/28 Thanh Thai, Phường 14, Quận 10, TPHCM
	switch len(arrFullAddress) {
	case 3:
		// Sample format: 7/28 Thanh Thai, Phường 4, Quận 10
		addressDetail.Street = arrFullAddress[0]
		addressDetail.Ward = arrFullAddress[1]
		addressDetail.District = arrFullAddress[2]
	case 4:
		//Case sample format:
		//1. 7/28 Thanh Thai, Quận 10, TPHCM, VN
		if listValueCountryVN[arrFullAddress[len(arrFullAddress)-1]] {
			addressDetail.Street = arrFullAddress[0]
			addressDetail.Ward = ""
			addressDetail.District = arrFullAddress[1]
			addressDetail.City = arrFullAddress[2]
			addressDetail.Country = arrFullAddress[3]
		} else {
			//Case sample format: 7/28 Thanh Thai, Phường 14, Quận 10, TPHCM
			addressDetail.Street = arrFullAddress[0]
			addressDetail.Ward = arrFullAddress[1]
			addressDetail.District = arrFullAddress[2]
			addressDetail.City = arrFullAddress[3]
		}
	case 5:
		addressDetail.Street = arrFullAddress[0]
		addressDetail.Ward = arrFullAddress[1]
		addressDetail.District = arrFullAddress[2]
		addressDetail.City = arrFullAddress[3]
		addressDetail.Country = arrFullAddress[4]
	}
	fmt.Println("XXX", addressDetail)
	return
}
