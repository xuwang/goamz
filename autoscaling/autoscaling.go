// The autoscaling package provides types and functions for interaction with the AWS
// AutoScaling service (autoscaling)
package autoscaling

import (
	"encoding/xml"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/goamz/aws"
)

// The AutoScaling type encapsulates operations operations with the autoscaling endpoint.
type AutoScaling struct {
	aws.Auth
	aws.Region
	httpClient *http.Client
}

const APIVersion = "2011-01-01"

// New creates a new AutoScaling instance.
func New(auth aws.Auth, region aws.Region) *AutoScaling {
	return NewWithClient(auth, region, aws.RetryingClient)
}

func NewWithClient(auth aws.Auth, region aws.Region, httpClient *http.Client) *AutoScaling {
	return &AutoScaling{auth, region, httpClient}
}

func (autoscaling *AutoScaling) query(params map[string]string, resp interface{}) error {
	params["Version"] = APIVersion
	params["Timestamp"] = time.Now().In(time.UTC).Format(time.RFC3339)

	endpoint, err := url.Parse(autoscaling.Region.AutoScalingEndpoint)
	if err != nil {
		return err
	}

	sign(autoscaling.Auth, "GET", "/", params, endpoint.Host)
	endpoint.RawQuery = multimap(params).Encode()
	r, err := autoscaling.httpClient.Get(endpoint.String())

	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode > 200 {
		return buildError(r)
	}

	decoder := xml.NewDecoder(r.Body)
	decodedBody := decoder.Decode(resp)

	return decodedBody
}

func buildError(r *http.Response) error {
	var (
		err    Error
		errors xmlErrors
	)
	xml.NewDecoder(r.Body).Decode(&errors)
	if len(errors.Errors) > 0 {
		err = errors.Errors[0]
	}
	err.StatusCode = r.StatusCode
	if err.Message == "" {
		err.Message = r.Status
	}
	return &err
}

func multimap(p map[string]string) url.Values {
	q := make(url.Values, len(p))
	for k, v := range p {
		q[k] = []string{v}
	}
	return q
}

func makeParams(action string) map[string]string {
	params := make(map[string]string)
	params["Action"] = action
	return params
}

// ----------------------------------------------------------------------------
// AutoScaling objects

type Tag struct {
	Key               string `xml:"Key"`
	Value             string `xml:"Value"`
	PropogateAtLaunch string `xml:"PropogateAtLaunch"`
}

type SecurityGroup struct {
	SecurityGroup string `xml:"member"`
}

type AvailabilityZone struct {
	AvailabilityZone string `xml:"member"`
}

type LoadBalancerName struct {
	LoadBalancerName string `xml:"member"`
}

type LaunchConfiguration struct {
	IamInstanceProfile string          `xml:"member>IamInstanceProfile"`
	ImageId            string          `xml:"member>ImageId"`
	InstanceType       string          `xml:"member>InstanceType"`
	KeyName            string          `xml:"member>KeyName"`
	Name               string          `xml:"member>LaunchConfigurationName"`
	SecurityGroups     []SecurityGroup `xml:"member>SecurityGroups"`
	UserData           []byte          `xml:"member>UserData"`
}

type AutoScalingGroup struct {
	AvailabilityZones       []AvailabilityZone `xml:"member>AvailabilityZones"`
	DefaultCooldown         int                `xml:"member>DefaultCooldown"`
	DesiredCapacity         int                `xml:"member>DesiredCapacity"`
	HealthCheckGracePeriod  int                `xml:"member>HealthCheckGracePeriod"`
	HealthCheckType         string             `xml:"member>HealthCheckType"`
	InstanceId              string             `xml:"member>InstanceId"`
	LaunchConfigurationName string             `xml:"member>LaunchConfigurationName"`
	LoadBalancerNames       []LoadBalancerName `xml:"member>LoadBalancerNames"`
	MaxSize                 int                `xml:"member>MaxSize"`
	MinSize                 int                `xml:"member>MinSize"`
	Name                    string             `xml:"member>AutoScalingGroupName"`
	VPCZoneIdentifier       string             `xml:"member>VPCZoneIdentifier"`
}

// ----------------------------------------------------------------------------
// Create

// The CreateAutoScalingGroup request parameters
type CreateAutoScalingGroup struct {
	AvailZone               []string
	DefaultCooldown         int
	DesiredCapacity         int
	HealthCheckGracePeriod  int
	HealthCheckType         string
	InstanceId              string
	LaunchConfigurationName string
	LoadBalancerNames       []string
	MaxSize                 int
	MinSize                 int
	PlacementGroup          string
	TerminationPolicies     []string
	Name                    string
	Tags                    []Tag
	VPCZoneIdentifier       []string

	SetDefaultCooldown        bool
	SetDesiredCapacity        bool
	SetHealthCheckGracePeriod bool
	SetMaxSize                bool
	SetMinSize                bool
}

func (autoscaling *AutoScaling) CreateAutoScalingGroup(options *CreateAutoScalingGroup) (resp *SimpleResp, err error) {
	params := makeParams("CreateAutoScalingGroup")

	params["AutoScalingGroupName"] = options.Name

	if options.SetDefaultCooldown {
		params["DefaultCooldown"] = strconv.Itoa(options.DefaultCooldown)
	}

	if options.SetDesiredCapacity {
		params["DesiredCapacity"] = strconv.Itoa(options.DesiredCapacity)
	}

	if options.SetHealthCheckGracePeriod {
		params["HealthCheckGracePeriod"] = strconv.Itoa(options.HealthCheckGracePeriod)
	}

	if options.HealthCheckType != "" {
		params["HealthCheckType"] = options.HealthCheckType
	}

	if options.InstanceId != "" {
		params["InstanceId"] = options.InstanceId
	}

	if options.LaunchConfigurationName != "" {
		params["LaunchConfigurationName"] = options.LaunchConfigurationName
	}

	for i, v := range options.AvailZone {
		params["AvailabilityZones.member."+strconv.Itoa(i+1)] = v
	}

	for i, v := range options.LoadBalancerNames {
		params["LoadBalancerNames.member."+strconv.Itoa(i+1)] = v
	}

	if options.SetMaxSize {
		params["MaxSize"] = strconv.Itoa(options.MaxSize)
	}

	if options.SetMinSize {
		params["MinSize"] = strconv.Itoa(options.MinSize)
	}

	if options.PlacementGroup != "" {
		params["PlacementGroup"] = options.PlacementGroup
	}

	for j, tag := range options.Tags {
		params["Tag.member."+strconv.Itoa(j+1)+".Key"] = tag.Key
		params["Tag.member."+strconv.Itoa(j+1)+".Value"] = tag.Value
	}

	for i, v := range options.TerminationPolicies {
		params["TerminationPolicies.member."+strconv.Itoa(i+1)] = v
	}

	if options.VPCZoneIdentifier != nil {
		params["VPCZoneIdentifier"] = strings.Join(options.VPCZoneIdentifier, ",")
	}

	resp = &SimpleResp{}

	err = autoscaling.query(params, resp)

	if err != nil {
		resp = nil
	}

	return
}

// The CreateLaunchConfiguration request parameters
type CreateLaunchConfiguration struct {
	IamInstanceProfile  string
	ImageId             string
	InstanceId          string
	InstanceType        string
	KeyName             string
	Name                string
	SecurityGroups      []string
	UserData            string
}

func (autoscaling *AutoScaling) CreateLaunchConfiguration(options *CreateLaunchConfiguration) (resp *SimpleResp, err error) {
	params := makeParams("CreateLaunchConfiguration")

	params["LaunchConfigurationName"] = options.Name
    
	if options.IamInstanceProfile != "" {
		params["IamInstanceProfile"] = options.IamInstanceProfile
	}
	if options.ImageId != "" {
		params["ImageId"] = options.ImageId
	}
	if options.InstanceType != "" {
		params["InstanceType"] = options.InstanceType
	}
	if options.InstanceId != "" {
		params["InstanceId"] = options.InstanceId
	}

	if options.KeyName != "" {
		params["KeyName"] = options.KeyName
	}

	for i, v := range options.SecurityGroups {
		params["SecurityGroups.member."+strconv.Itoa(i+1)] = v
	}

	if options.UserData != "" {
		userData := make([]byte, b64.EncodedLen(len(options.UserData)))
		b64.Encode(userData, []byte(options.UserData))
		params["UserData"] = string(userData)
	}

	resp = &SimpleResp{}

	err = autoscaling.query(params, resp)

	if err != nil {
		resp = nil
	}

	return
}

// Describe

// DescribeAutoScalingGroups request params
type DescribeAutoScalingGroups struct {
	Names []string
}

type DescribeAutoScalingGroupsResp struct {
	RequestId         string             `xml:"ResponseMetadata>RequestId"`
	AutoScalingGroups []AutoScalingGroup `xml:"DescribeAutoScalingGroupsResult>AutoScalingGroups"`
}

func (autoscaling *AutoScaling) DescribeAutoScalingGroups(options *DescribeAutoScalingGroups) (resp *DescribeAutoScalingGroupsResp, err error) {
	params := makeParams("DescribeAutoScalingGroups")

	for i, v := range options.Names {
		params["AutoScalingGroupNames.member."+strconv.Itoa(i+1)] = v
	}

	resp = &DescribeAutoScalingGroupsResp{}

	err = autoscaling.query(params, resp)

	if err != nil {
		resp = nil
	}

	return
}

// DescribeLaunchConfigurations request params
type DescribeLaunchConfigurations struct {
	Names []string
}

type DescribeLaunchConfigurationsResp struct {
	RequestId            string                `xml:"ResponseMetadata>RequestId"`
	LaunchConfigurations []LaunchConfiguration `xml:"DescribeLaunchConfigurationsResult>LaunchConfigurations"`
}

func (autoscaling *AutoScaling) DescribeLaunchConfigurations(options *DescribeLaunchConfigurations) (resp *DescribeLaunchConfigurationsResp, err error) {
	params := makeParams("DescribeLaunchConfigurations")

	for i, v := range options.Names {
		params["LaunchConfigurationNames.member."+strconv.Itoa(i+1)] = v
	}

	resp = &DescribeLaunchConfigurationsResp{}

	err = autoscaling.query(params, resp)

	if err != nil {
		resp = nil
	}

	return
}

// ----------------------------------------------------------------------------
// Destroy

// The DeleteLaunchConfiguration request parameters
type DeleteLaunchConfiguration struct {
	Name string
}

func (autoscaling *AutoScaling) DeleteLaunchConfiguration(options *DeleteLaunchConfiguration) (resp *SimpleResp, err error) {
	params := makeParams("DeleteLaunchConfiguration")

	params["LaunchConfigurationName"] = options.Name

	resp = &SimpleResp{}

	err = autoscaling.query(params, resp)

	if err != nil {
		resp = nil
	}

	return
}

// The DeleteLaunchConfiguration request parameters
type DeleteAutoScalingGroup struct {
	Name        string
	ForceDelete bool
}

func (autoscaling *AutoScaling) DeleteAutoScalingGroup(options *DeleteAutoScalingGroup) (resp *SimpleResp, err error) {
	params := makeParams("DeleteAutoScalingGroup")

	params["AutoScalingGroupName"] = options.Name
	params["ForceDelete"] = strconv.FormatBool(options.ForceDelete)

	resp = &SimpleResp{}

	err = autoscaling.query(params, resp)

	if err != nil {
		resp = nil
	}

	return
}

// ----------------------------------------------------------------------------
// Destroy
// The UpdateAutoScalingGroup request parameters
type UpdateAutoScalingGroup struct {
	AvailZone               []string
	DefaultCooldown         int
	DesiredCapacity         int
	HealthCheckGracePeriod  int
	HealthCheckType         string
	LaunchConfigurationName string
	MaxSize                 int
	MinSize                 int
	PlacementGroup          string
	TerminationPolicies     []string
	Name                    string
	VPCZoneIdentifier       []string

	SetDefaultCooldown        bool
	SetDesiredCapacity        bool
	SetHealthCheckGracePeriod bool
	SetMaxSize                bool
	SetMinSize                bool
}

func (autoscaling *AutoScaling) UpdateAutoScalingGroup(options *UpdateAutoScalingGroup) (resp *SimpleResp, err error) {
	params := makeParams("UpdateAutoScalingGroup")

	if options.Name != "" {
		params["AutoScalingGroupName"] = options.Name
	}

	if options.SetDefaultCooldown {
		params["DefaultCooldown"] = strconv.Itoa(options.DefaultCooldown)
	}

	if options.SetDesiredCapacity {
		params["DesiredCapacity"] = strconv.Itoa(options.DesiredCapacity)
	}

	if options.SetHealthCheckGracePeriod {
		params["HealthCheckGracePeriod"] = strconv.Itoa(options.HealthCheckGracePeriod)
	}

	if options.HealthCheckType != "" {
		params["HealthCheckType"] = options.HealthCheckType
	}

	if options.LaunchConfigurationName != "" {
		params["LaunchConfigurationName"] = options.LaunchConfigurationName
	}

	for i, v := range options.AvailZone {
		params["AvailabilityZones.member."+strconv.Itoa(i+1)] = v
	}

	if options.SetMaxSize {
		params["MaxSize"] = strconv.Itoa(options.MaxSize)
	}

	if options.SetMinSize {
		params["MinSize"] = strconv.Itoa(options.MinSize)
	}

	if options.PlacementGroup != "" {
		params["PlacementGroup"] = options.PlacementGroup
	}
	for i, v := range options.TerminationPolicies {
		params["TerminationPolicies.member."+strconv.Itoa(i+1)] = v
	}

	if options.VPCZoneIdentifier != nil {
		params["VPCZoneIdentifier"] = strings.Join(options.VPCZoneIdentifier, ",")
	}

	resp = &SimpleResp{}

	err = autoscaling.query(params, resp)

	if err != nil {
		resp = nil
	}

	return
}

// Responses

type SimpleResp struct {
	RequestId string `xml:"ResponseMetadata>RequestId"`
}

type xmlErrors struct {
	Errors []Error `xml:"Error"`
}

// Error encapsulates an autoscaling error.
type Error struct {
	// HTTP status code of the error.
	StatusCode int

	// AWS code of the error.
	Code string

	// Message explaining the error.
	Message string
}

func (e *Error) Error() string {
	var prefix string
	if e.Code != "" {
		prefix = e.Code + ": "
	}
	if prefix == "" && e.StatusCode > 0 {
		prefix = strconv.Itoa(e.StatusCode) + ": "
	}
	return prefix + e.Message
}
