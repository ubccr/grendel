// This file is auto-generated by @hey-api/openapi-ts

import { createClient, createConfig, type Options } from '@hey-api/client-fetch';
import type { HostListError, HostListResponse, HostFindData, HostFindError, HostFindResponse, HostDeleteData, HostDeleteError, HostDeleteResponse, HostTagsData, HostTagsError, HostTagsResponse, HostProvisionData, HostProvisionError, HostProvisionResponse, HostUnprovisionData, HostUnprovisionError, HostUnprovisionResponse, HostTagData, HostTagError, HostTagResponse, HostUntagData, HostUntagError, HostUntagResponse, StoreHostsData, StoreHostsError, StoreHostsResponse, UserListError, UserListResponse, ImageListError, ImageListResponse, ImageFindData, ImageFindError, ImageFindResponse, ImageDeleteData, ImageDeleteError, ImageDeleteResponse, StoreImagesData, StoreImagesError, StoreImagesResponse, RestoreData, RestoreError, RestoreResponse } from './types.gen';

export const client = createClient(createConfig());

/**
 * List all hosts
 * Returns all host stored in Grendel
 */
export const hostList = <ThrowOnError extends boolean = false>(options?: Options<unknown, ThrowOnError>) => { return (options?.client ?? client).get<HostListResponse, HostListError, ThrowOnError>({
    ...options,
    url: '/host/list'
}); };

/**
 * Find hosts by name or nodeset
 * Returns a list of hosts in the given nodeset
 */
export const hostFind = <ThrowOnError extends boolean = false>(options: Options<HostFindData, ThrowOnError>) => { return (options?.client ?? client).get<HostFindResponse, HostFindError, ThrowOnError>({
    ...options,
    url: '/host/find/{nodeSet}'
}); };

/**
 * Delete hosts by name or nodeset
 * Delete hosts in the given nodeset
 */
export const hostDelete = <ThrowOnError extends boolean = false>(options: Options<HostDeleteData, ThrowOnError>) => { return (options?.client ?? client).delete<HostDeleteResponse, HostDeleteError, ThrowOnError>({
    ...options,
    url: '/host/find/{nodeSet}'
}); };

/**
 * Find hosts by tags
 * Returns a list of hosts with the given tags
 */
export const hostTags = <ThrowOnError extends boolean = false>(options: Options<HostTagsData, ThrowOnError>) => { return (options?.client ?? client).get<HostTagsResponse, HostTagsError, ThrowOnError>({
    ...options,
    url: '/host/tags/{tags}'
}); };

/**
 * Set hosts to provision by name or nodeset
 * Sets hosts to provision in the given nodeset
 */
export const hostProvision = <ThrowOnError extends boolean = false>(options: Options<HostProvisionData, ThrowOnError>) => { return (options?.client ?? client).put<HostProvisionResponse, HostProvisionError, ThrowOnError>({
    ...options,
    url: '/host/provision/{nodeSet}'
}); };

/**
 * Set hosts to unprovision by name or nodeset
 * Sets hosts to unprovision in the given nodeset
 */
export const hostUnprovision = <ThrowOnError extends boolean = false>(options: Options<HostUnprovisionData, ThrowOnError>) => { return (options?.client ?? client).put<HostUnprovisionResponse, HostUnprovisionError, ThrowOnError>({
    ...options,
    url: '/host/unprovision/{nodeSet}'
}); };

/**
 * Tag hosts by name or nodeset
 * Tag hosts in the given nodeset
 */
export const hostTag = <ThrowOnError extends boolean = false>(options: Options<HostTagData, ThrowOnError>) => { return (options?.client ?? client).put<HostTagResponse, HostTagError, ThrowOnError>({
    ...options,
    url: '/host/tag/{nodeSet}'
}); };

/**
 * Untag hosts name or nodeset
 * Untag hosts the given nodeset
 */
export const hostUntag = <ThrowOnError extends boolean = false>(options: Options<HostUntagData, ThrowOnError>) => { return (options?.client ?? client).put<HostUntagResponse, HostUntagError, ThrowOnError>({
    ...options,
    url: '/host/untag/{nodeSet}'
}); };

/**
 * Add or update hosts in Grendel
 */
export const storeHosts = <ThrowOnError extends boolean = false>(options: Options<StoreHostsData, ThrowOnError>) => { return (options?.client ?? client).post<StoreHostsResponse, StoreHostsError, ThrowOnError>({
    ...options,
    url: '/host'
}); };

/**
 * List all users
 * Returns all users stored in Grendel
 */
export const userList = <ThrowOnError extends boolean = false>(options?: Options<unknown, ThrowOnError>) => { return (options?.client ?? client).get<UserListResponse, UserListError, ThrowOnError>({
    ...options,
    url: '/user/list'
}); };

/**
 * List all images
 * Returns all images stored in Grendel
 */
export const imageList = <ThrowOnError extends boolean = false>(options?: Options<unknown, ThrowOnError>) => { return (options?.client ?? client).get<ImageListResponse, ImageListError, ThrowOnError>({
    ...options,
    url: '/bootimage/list'
}); };

/**
 * Find image by name
 * Returns image with the given name
 */
export const imageFind = <ThrowOnError extends boolean = false>(options: Options<ImageFindData, ThrowOnError>) => { return (options?.client ?? client).get<ImageFindResponse, ImageFindError, ThrowOnError>({
    ...options,
    url: '/bootimage/find/{name}'
}); };

/**
 * Delete boot images by name
 * Delete boot images with the given name
 */
export const imageDelete = <ThrowOnError extends boolean = false>(options: Options<ImageDeleteData, ThrowOnError>) => { return (options?.client ?? client).delete<ImageDeleteResponse, ImageDeleteError, ThrowOnError>({
    ...options,
    url: '/bootimage/find/{name}'
}); };

/**
 * Add or update images in Grendel
 */
export const storeImages = <ThrowOnError extends boolean = false>(options: Options<StoreImagesData, ThrowOnError>) => { return (options?.client ?? client).post<StoreImagesResponse, StoreImagesError, ThrowOnError>({
    ...options,
    url: '/bootimage'
}); };

/**
 * Restore Grendel database
 */
export const restore = <ThrowOnError extends boolean = false>(options: Options<RestoreData, ThrowOnError>) => { return (options?.client ?? client).post<RestoreResponse, RestoreError, ThrowOnError>({
    ...options,
    url: '/restore'
}); };