package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/brycekahle/goamz/ec2"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/elb"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type csv []string

func (i *csv) String() string {
	return fmt.Sprint(*i)
}

func (i *csv) Set(value string) error {
	if len(*i) > 0 {
		return errors.New("csv flag already set")
	}
	for _, dt := range strings.Split(value, ",") {
		*i = append(*i, dt)
	}
	return nil
}

var region, accesskey, secretkey, securityGroupID string
var lbnames csv
var awsec2 *ec2.EC2

func init() {
	flag.Var(&lbnames, "lbname", "name of the ELB to register the instance with")
	flag.StringVar(&securityGroupID, "groupid", os.Getenv("SECURITY_GROUP_ID"), "id of the EC2 security group to add this instance to")
	flag.StringVar(&region, "region", os.Getenv("AWS_REGION"), "AWS region in which the ELB resides")
	flag.StringVar(&accesskey, "accesskey", os.Getenv("AWS_ACCESS_KEY"), "AWS Access Key")
	flag.StringVar(&secretkey, "secretkey", os.Getenv("AWS_SECRET_KEY"), "AWS Secret Key")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "go-elb-presence\n")
		flag.PrintDefaults()
	}

	flag.Parse()
	if len(lbnames) == 0 {
		lbnames.Set(os.Getenv("ELB_NAME"))
	}
}

func main() {
	instanceID := aws.InstanceId()
	if instanceID == "unknown" {
		log.Fatalln("Unable to get instance id")
	}

	auth, err := aws.GetAuth(accesskey, secretkey, "", time.Time{})
	if err != nil {
		log.Fatalln("Unable to get AWS auth", err)
	}

	if securityGroupID != "" {
		awsec2 = ec2.New(auth, aws.GetRegion(region))

		groupMap := getSecurityGroupIds(instanceID)
		groupMap[securityGroupID] = true
		groupIds := make([]string, 0, len(groupMap))
		for id := range groupMap {
			groupIds = append(groupIds, id)
		}

		opts := &ec2.ModifyInstanceAttributeOptions{SecurityGroups: ec2.SecurityGroupIds(groupIds...)}
		resp, err := awsec2.ModifyInstanceAttribute(instanceID, opts)
		if err != nil || !resp.Return {
			log.Fatalln("Error adding security group to instance", err)
		}

		log.Printf("Added security group %s to instance %s\n", securityGroupID, instanceID)
	}

	awselb := elb.New(auth, aws.GetRegion(region))
	for _, lbname := range lbnames {
		_, err = awselb.RegisterInstancesWithLoadBalancer([]string{instanceID}, lbname)
		if err != nil {
			log.Fatalln("Error registering instance", err)
		}

		log.Printf("Registered instance %s with elb %s\n", instanceID, lbname)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	// this waits until we get a kill signal
	<-c

	for _, lbname := range lbnames {
		_, err = awselb.DeregisterInstancesFromLoadBalancer([]string{instanceID}, lbname)
		if err != nil {
			log.Fatalln("Error deregistering instance", err)
		}

		log.Printf("Deregistered instance %s with elb %s\n", instanceID, lbname)
	}

	if securityGroupID != "" {
		groupMap := getSecurityGroupIds(instanceID)
		delete(groupMap, securityGroupID)
		groupIds := make([]string, 0, len(groupMap))
		for id := range groupMap {
			groupIds = append(groupIds, id)
		}

		opts := &ec2.ModifyInstanceAttributeOptions{SecurityGroups: ec2.SecurityGroupIds(groupIds...)}
		resp, err := awsec2.ModifyInstanceAttribute(instanceID, opts)
		if err != nil || !resp.Return {
			log.Fatalln("Error removing security group from instance", err)
		}

		log.Printf("Removed security group %s from instance %s\n", securityGroupID, instanceID)
	}
}

func getSecurityGroupIds(instanceID string) map[string]bool {
	resp, err := awsec2.DescribeInstanceAttribute(instanceID, "groupSet")
	if err != nil {
		log.Fatalln("Error describing instance attributes", err)
	}

	groupIds := make(map[string]bool)
	for _, g := range resp.SecurityGroups {
		groupIds[g.Id] = true
	}
	return groupIds
}
