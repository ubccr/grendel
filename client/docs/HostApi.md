# \HostApi

All URIs are relative to *http://localhost/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**HostFind**](HostApi.md#HostFind) | **Get** /host/find/{nodeSet} | Find hosts by name or nodeset
[**HostList**](HostApi.md#HostList) | **Get** /host/list | List all hosts
[**StoreHosts**](HostApi.md#StoreHosts) | **Post** /host | Add or update hosts in Grendel



## HostFind

> []Host HostFind(ctx, nodeSet)

Find hosts by name or nodeset

Returns a list of hosts in the given nodeset

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**nodeSet** | **string**| nodeset syntax. Example: cpn-d13-[01-100] | 

### Return type

[**[]Host**](Host.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## HostList

> []Host HostList(ctx, )

List all hosts

Returns all host stored in Grendel

### Required Parameters

This endpoint does not need any parameter.

### Return type

[**[]Host**](Host.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## StoreHosts

> StoreHosts(ctx, body)

Add or update hosts in Grendel

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**body** | [**[]Host**](Host.md)| Add/Update one or more hosts to Grendel | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

