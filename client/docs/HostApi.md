# \HostApi

All URIs are relative to *http://localhost/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**HostFind**](HostApi.md#HostFind) | **Get** /host/find/{nodeSet} | Find hosts by name or nodeset
[**HostList**](HostApi.md#HostList) | **Get** /host/list | List all hosts
[**HostProvision**](HostApi.md#HostProvision) | **Put** /host/provision/{nodeSet} | Set hosts to provision by name or nodeset
[**HostTag**](HostApi.md#HostTag) | **Put** /host/tag/{nodeSet} | Tag hosts by name or nodeset
[**HostTags**](HostApi.md#HostTags) | **Get** /host/tags/{tags} | Find hosts by tags
[**HostUnprovision**](HostApi.md#HostUnprovision) | **Put** /host/unprovision/{nodeSet} | Set hosts to unprovision by name or nodeset
[**HostUntag**](HostApi.md#HostUntag) | **Put** /host/untag/{nodeSet} | Untag hosts name or nodeset
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


## HostProvision

> HostProvision(ctx, nodeSet)

Set hosts to provision by name or nodeset

Sets hosts to provision in the given nodeset

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**nodeSet** | **string**| nodeset syntax. Example: cpn-d13-[01-100] | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## HostTag

> HostTag(ctx, nodeSet, tags)

Tag hosts by name or nodeset

Tag hosts in the given nodeset

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**nodeSet** | **string**| nodeset syntax. Example: cpn-d13-[01-100] | 
**tags** | **string**| list of tags. Example: ib,noib,k16 | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## HostTags

> []Host HostTags(ctx, tags)

Find hosts by tags

Returns a list of hosts with the given tags

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tags** | **string**| comman separated list of tags. Example: ib,noib,k16 | 

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


## HostUnprovision

> HostUnprovision(ctx, nodeSet)

Set hosts to unprovision by name or nodeset

Sets hosts to unprovision in the given nodeset

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**nodeSet** | **string**| nodeset syntax. Example: cpn-d13-[01-100] | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## HostUntag

> HostUntag(ctx, nodeSet, tags)

Untag hosts name or nodeset

Untag hosts the given nodeset

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**nodeSet** | **string**| nodeset syntax. Example: cpn-d13-[01-100] | 
**tags** | **string**| list of tags. Example: ib,noib,k16 | 

### Return type

 (empty response body)

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

