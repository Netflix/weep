package util

import (
	"errors"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

type AwsArn struct {
	Arn               string
	Partition         string
	Service           string
	Region            string
	AccountId         string
	ResourceType      string
	Resource          string
	ResourceDelimiter string
}


func validate(arn string, pieces []string) error {
	if len(pieces) < 6 {
		return errors.New("Malformed ARN")
	}
	return nil
}

func ArnParse(arn string) (*AwsArn, error) {
	pieces := strings.SplitN(arn, ":", 6)

	if err := validate(arn, pieces); err != nil {
		return nil, err
	}

	components := &AwsArn{
		Arn:       pieces[0],
		Partition: pieces[1],
		Service:   pieces[2],
		Region:    pieces[3],
		AccountId: pieces[4],
	}
	if n := strings.Count(pieces[5], ":"); n > 0 {
		components.ResourceDelimiter = ":"
		resourceParts := strings.SplitN(pieces[5], ":", 2)
		components.ResourceType = resourceParts[0]
		components.Resource = resourceParts[1]
	} else {
		if m := strings.Count(pieces[5], "/"); m == 0 {
			components.Resource = pieces[5]
		} else {
			components.ResourceDelimiter = "/"
			resourceParts := strings.SplitN(pieces[5], "/", 2)
			components.ResourceType = resourceParts[0]
			components.Resource = resourceParts[1]
		}
	}
	return components, nil
}

func (a AwsArn) ArnString() string {
	return fmt.Sprintf("%s:%s:%s:%s:%s:%s/%s", a.Arn, a.Partition, a.Service, a.Region, a.AccountId, a.ResourceType, a.Resource)
}

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
