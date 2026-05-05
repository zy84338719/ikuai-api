package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	ikuaiapi "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/types"
)

type firewallService struct {
	client *ikuaiapi.Client
}

func NewFirewallService(client *ikuaiapi.Client) FirewallService {
	return &firewallService{client: client}
}

func (s *firewallService) GetACL(ctx context.Context) ([]types.ACLItem, error) {
	var resp types.ACLShowResponse
	if err := s.client.Call(ctx, "acl", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *firewallService) GetDNAT(ctx context.Context) ([]types.DNATItem, error) {
	var resp types.DNATShowResponse
	if err := s.client.Call(ctx, "dnat", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *firewallService) GetConnLimit(ctx context.Context) ([]types.ConnLimitItem, error) {
	var resp types.ConnLimitShowResponse
	if err := s.client.Call(ctx, "conn_limit", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *firewallService) GetDomainGroups(ctx context.Context) ([]types.DomainGroupItem, error) {
	var resp types.DomainGroupShowResponse
	if err := s.client.Call(ctx, "domain_group", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *firewallService) GetCustomISP(ctx context.Context) ([]types.CustomISPItem, error) {
	var resp types.CustomISPShowResponse
	if err := s.client.Call(ctx, "custom_isp", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *firewallService) GetStreamDomain(ctx context.Context) ([]types.StreamDomainItem, error) {
	var resp types.StreamDomainShowResponse
	if err := s.client.Call(ctx, "stream_domain", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *firewallService) AddACL(ctx context.Context, req *types.ACLAddRequest) (int, error) {
	var result struct {
		types.BaseResponse
		ID int `json:"id"`
	}
	if err := s.client.Call(ctx, "acl", "add", req, &result); err != nil {
		return 0, err
	}
	if !result.IsSuccess() {
		return 0, fmt.Errorf("failed to add ACL: %s", result.GetErrorMessage())
	}
	return result.ID, nil
}

func (s *firewallService) EditACL(ctx context.Context, req *types.ACLEditRequest) error {
	var result types.BaseResponse
	if err := s.client.Call(ctx, "acl", "edit", req, &result); err != nil {
		return err
	}
	if !result.IsSuccess() {
		return fmt.Errorf("failed to edit ACL: %s", result.GetErrorMessage())
	}
	return nil
}

func (s *firewallService) DelACL(ctx context.Context, id int) error {
	var result types.BaseResponse
	if err := s.client.Call(ctx, "acl", "del", map[string]int{"id": id}, &result); err != nil {
		return err
	}
	if !result.IsSuccess() {
		return fmt.Errorf("failed to delete ACL: %s", result.GetErrorMessage())
	}
	return nil
}

func (s *firewallService) AddDNAT(ctx context.Context, req *types.DNATAddRequest) (int, error) {
	var result struct {
		types.BaseResponse
		ID int `json:"id"`
	}
	if err := s.client.Call(ctx, "dnat", "add", req, &result); err != nil {
		return 0, err
	}
	if !result.IsSuccess() {
		return 0, fmt.Errorf("failed to add DNAT: %s", result.GetErrorMessage())
	}
	return result.ID, nil
}

func (s *firewallService) EditDNAT(ctx context.Context, req *types.DNATEditRequest) error {
	var result types.BaseResponse
	if err := s.client.Call(ctx, "dnat", "edit", req, &result); err != nil {
		return err
	}
	if !result.IsSuccess() {
		return fmt.Errorf("failed to edit DNAT: %s", result.GetErrorMessage())
	}
	return nil
}

func (s *firewallService) DelDNAT(ctx context.Context, id int) error {
	var result types.BaseResponse
	if err := s.client.Call(ctx, "dnat", "del", map[string]int{"id": id}, &result); err != nil {
		return err
	}
	if !result.IsSuccess() {
		return fmt.Errorf("failed to delete DNAT: %s", result.GetErrorMessage())
	}
	return nil
}

func (s *firewallService) AddConnLimit(ctx context.Context, req *types.ConnLimitAddRequest) (int, error) {
	var result struct {
		types.BaseResponse
		ID int `json:"id"`
	}
	if err := s.client.Call(ctx, "conn_limit", "add", req, &result); err != nil {
		return 0, err
	}
	if !result.IsSuccess() {
		return 0, fmt.Errorf("failed to add conn_limit: %s", result.GetErrorMessage())
	}
	return result.ID, nil
}

func (s *firewallService) EditConnLimit(ctx context.Context, req *types.ConnLimitEditRequest) error {
	var result types.BaseResponse
	if err := s.client.Call(ctx, "conn_limit", "edit", req, &result); err != nil {
		return err
	}
	if !result.IsSuccess() {
		return fmt.Errorf("failed to edit conn_limit: %s", result.GetErrorMessage())
	}
	return nil
}

func (s *firewallService) DelConnLimit(ctx context.Context, id int) error {
	var result types.BaseResponse
	if err := s.client.Call(ctx, "conn_limit", "del", map[string]int{"id": id}, &result); err != nil {
		return err
	}
	if !result.IsSuccess() {
		return fmt.Errorf("failed to delete conn_limit: %s", result.GetErrorMessage())
	}
	return nil
}

func (s *firewallService) AddCustomISP(ctx context.Context, name string, ipGroups []string, comment string) (int, error) {
	if comment == "" {
		comment = "ikuai-aio"
	}
	// deduplicate
	seen := make(map[string]struct{}, len(ipGroups))
	unique := ipGroups[:0]
	for _, ip := range ipGroups {
		if _, ok := seen[ip]; !ok {
			seen[ip] = struct{}{}
			unique = append(unique, ip)
		}
	}

	const chunkSize = 5000
	total := 0
	for i := 0; i < len(unique); i += chunkSize {
		end := i + chunkSize
		if end > len(unique) {
			end = len(unique)
		}
		chunk := unique[i:end]
		req := &types.CustomISPAddRequest{
			Name:    name,
			IPGroup: strings.Join(chunk, ","),
			Comment: comment,
		}
		var result types.BaseResponse
		if err := s.client.Call(ctx, "custom_isp", "add", req, &result); err != nil {
			return total, err
		}
		if !result.IsSuccess() {
			return total, fmt.Errorf("failed to add custom_isp: %s", result.GetErrorMessage())
		}
		total += len(chunk)
	}
	return total, nil
}

func (s *firewallService) DelCustomISP(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	strs := make([]string, len(ids))
	for i, id := range ids {
		strs[i] = strconv.Itoa(id)
	}
	req := &types.CustomISPDelRequest{ID: strings.Join(strs, ",")}
	var result types.BaseResponse
	if err := s.client.Call(ctx, "custom_isp", "del", req, &result); err != nil {
		return err
	}
	if !result.IsSuccess() {
		return fmt.Errorf("failed to delete custom_isp: %s", result.GetErrorMessage())
	}
	return nil
}

func (s *firewallService) AddStreamDomain(ctx context.Context, interfaces []string, domains []string, srcAddr, comment string) (int, error) {
	if comment == "" {
		comment = "ikuai-aio"
	}
	// deduplicate
	seen := make(map[string]struct{}, len(domains))
	unique := domains[:0]
	for _, d := range domains {
		if _, ok := seen[d]; !ok {
			seen[d] = struct{}{}
			unique = append(unique, d)
		}
	}

	const chunkSize = 1000
	total := 0
	for i := 0; i < len(unique); i += chunkSize {
		end := i + chunkSize
		if end > len(unique) {
			end = len(unique)
		}
		chunk := unique[i:end]
		req := &types.StreamDomainAddRequest{
			Interface: strings.Join(interfaces, ","),
			SrcAddr:   srcAddr,
			Domain:    strings.Join(chunk, ","),
			Comment:   comment,
			Week:      "1234567",
			Time:      "00:00-23:59",
			Enabled:   "yes",
		}
		var result types.BaseResponse
		if err := s.client.Call(ctx, "stream_domain", "add", req, &result); err != nil {
			return total, err
		}
		if !result.IsSuccess() {
			return total, fmt.Errorf("failed to add stream_domain: %s", result.GetErrorMessage())
		}
		total += len(chunk)
	}
	return total, nil
}

func (s *firewallService) DelStreamDomain(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	strs := make([]string, len(ids))
	for i, id := range ids {
		strs[i] = strconv.Itoa(id)
	}
	req := &types.StreamDomainDelRequest{ID: strings.Join(strs, ",")}
	var result types.BaseResponse
	if err := s.client.Call(ctx, "stream_domain", "del", req, &result); err != nil {
		return err
	}
	if !result.IsSuccess() {
		return fmt.Errorf("failed to delete stream_domain: %s", result.GetErrorMessage())
	}
	return nil
}
