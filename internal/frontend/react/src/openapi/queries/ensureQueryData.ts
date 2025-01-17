// generated with @7nohe/openapi-react-query-codegen@2.0.0-beta.3 

import { type Options } from "@hey-api/client-fetch";
import { type QueryClient } from "@tanstack/react-query";
import { hostFind, hostList, hostTags, imageFind, imageList, userList } from "../requests/services.gen";
import { HostFindData, HostTagsData, ImageFindData } from "../requests/types.gen";
import * as Common from "./common";
export const ensureUseHostListData = (queryClient: QueryClient, clientOptions: Options<unknown, true> = {}) => queryClient.ensureQueryData({ queryKey: Common.UseHostListKeyFn(clientOptions), queryFn: () => hostList({ ...clientOptions }).then(response => response.data) });
export const ensureUseHostFindData = (queryClient: QueryClient, clientOptions: Options<HostFindData, true>) => queryClient.ensureQueryData({ queryKey: Common.UseHostFindKeyFn(clientOptions), queryFn: () => hostFind({ ...clientOptions }).then(response => response.data) });
export const ensureUseHostTagsData = (queryClient: QueryClient, clientOptions: Options<HostTagsData, true>) => queryClient.ensureQueryData({ queryKey: Common.UseHostTagsKeyFn(clientOptions), queryFn: () => hostTags({ ...clientOptions }).then(response => response.data) });
export const ensureUseUserListData = (queryClient: QueryClient, clientOptions: Options<unknown, true> = {}) => queryClient.ensureQueryData({ queryKey: Common.UseUserListKeyFn(clientOptions), queryFn: () => userList({ ...clientOptions }).then(response => response.data) });
export const ensureUseImageListData = (queryClient: QueryClient, clientOptions: Options<unknown, true> = {}) => queryClient.ensureQueryData({ queryKey: Common.UseImageListKeyFn(clientOptions), queryFn: () => imageList({ ...clientOptions }).then(response => response.data) });
export const ensureUseImageFindData = (queryClient: QueryClient, clientOptions: Options<ImageFindData, true>) => queryClient.ensureQueryData({ queryKey: Common.UseImageFindKeyFn(clientOptions), queryFn: () => imageFind({ ...clientOptions }).then(response => response.data) });
