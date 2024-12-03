# \ImageApi

All URIs are relative to *http://localhost/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ImageDelete**](ImageApi.md#ImageDelete) | **Delete** /bootimage/find/{name} | Delete boot images by name
[**ImageFind**](ImageApi.md#ImageFind) | **Get** /bootimage/find/{name} | Find image by name
[**ImageList**](ImageApi.md#ImageList) | **Get** /bootimage/list | List all images
[**StoreImages**](ImageApi.md#StoreImages) | **Post** /bootimage | Add or update images in Grendel



## ImageDelete

> ImageDelete(ctx, name)

Delete boot images by name

Delete boot images with the given name

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**name** | **string**| Name of boot image to delete | 

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


## ImageFind

> []BootImage ImageFind(ctx, name)

Find image by name

Returns image with the given name

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**name** | **string**| Name of image | 

### Return type

[**[]BootImage**](BootImage.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ImageList

> []BootImage ImageList(ctx, )

List all images

Returns all images stored in Grendel

### Required Parameters

This endpoint does not need any parameter.

### Return type

[**[]BootImage**](BootImage.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## StoreImages

> StoreImages(ctx, body)

Add or update images in Grendel

### Required Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**body** | [**[]BootImage**](BootImage.md)| Add/Update one or more images to Grendel | 

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

