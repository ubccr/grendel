import type { Event } from "@/client";
import {
  getV1BmcBiosOptions,
  getV1BmcFirmwareUpdatesOptions,
  getV1BmcJobsOptions,
  getV1BmcSecurebootOptions,
  getV1BmcSelOptions,
  getV1BmcStatusOptions,
  getV1GrendelEventsOptions,
  getV1NodesCheckOptions,
  getV1NodesFindOptions,
  getV1NodesProvisionStatusOptions,
  getV1NodesRenderTemplateOptions,
  getV1SwitchNodesetLldpOptions,
  getV1SwitchScanOptions,
} from "@/client/@tanstack/react-query.gen";
import { isRunning, newestFirst } from "@/lib/events";
import { type Query, queryOptions, useQueryClient } from "@tanstack/react-query";
import { useCallback } from "react";

// a route's loader and the components under it take their options from here, so both read the one cache entry

// what a BMC or switch said is kept longer than a read of Grendel's database, since asking again logs into the device and takes seconds
const hardware = {
  staleTime: 30_000,
  // a device that did not answer is not asked again just because another component mounted
  retryOnMount: false,
} as const;

export const nodeOptions = (node: string) => getV1NodesFindOptions({ query: { nodeset: node } });

export const provisionStatusOptions = (query?: { nodeset?: string; tags?: string }) =>
  getV1NodesProvisionStatusOptions(query && { query });

// both read the templates and boot files from the server's disk, which is what changes between looks, so neither is served from the cache
export const bootCheckOptions = (nodes: string, arch?: string) =>
  queryOptions({ ...getV1NodesCheckOptions({ query: { nodeset: nodes, arch } }), staleTime: 0 });

export const renderOptions = (node: string, template: string, arch?: string) =>
  queryOptions({
    ...getV1NodesRenderTemplateOptions({ path: { template }, query: { nodeset: node, arch } }),
    staleTime: 0,
    // the output holds a working boot token, which is dropped as soon as nothing shows it
    gcTime: 0,
  });

export const bmcStatusOptions = (nodes: string) =>
  queryOptions({ ...getV1BmcStatusOptions({ query: { nodeset: nodes } }), ...hardware });

export const biosOptions = (nodes: string) =>
  queryOptions({ ...getV1BmcBiosOptions({ query: { nodeset: nodes } }), ...hardware });

// one request reads the state and the keys over a single login to each BMC
export const secureBootOptions = (nodes: string) =>
  queryOptions({ ...getV1BmcSecurebootOptions({ query: { nodeset: nodes } }), ...hardware });

// a BMC hands over each key in a request of its own, so the few in PK, KEK and db are read apart from the hundreds of hashes in dbx
export const secureBootKeySets = { trusted: "PK,KEK,db", revoked: "dbx" } as const;

export const secureBootKeysOptions = (nodes: string, set: keyof typeof secureBootKeySets) =>
  queryOptions({
    ...getV1BmcSecurebootOptions({
      query: { nodeset: nodes, keys: true, databases: secureBootKeySets[set] },
    }),
    ...hardware,
  });

// limit is how many entries to read from each BMC, with 0 for the whole log
export const selOptions = (nodes: string, limit: number) =>
  queryOptions({ ...getV1BmcSelOptions({ query: { nodeset: nodes, limit } }), ...hardware });

export const jobsOptions = (nodes: string) =>
  queryOptions({ ...getV1BmcJobsOptions({ query: { nodeset: nodes } }), ...hardware });

export const firmwareUpdatesOptions = (nodes: string) =>
  queryOptions({ ...getV1BmcFirmwareUpdatesOptions({ query: { nodeset: nodes } }), ...hardware });

export const lldpOptions = (node: string) =>
  queryOptions({ ...getV1SwitchNodesetLldpOptions({ path: { nodeset: node } }), ...hardware });

export const switchScanOptions = (node: string) =>
  queryOptions({ ...getV1SwitchScanOptions({ query: { nodeset: node } }), ...hardware });

// EVENTS_SHOWN is how many events are asked for, the most recently started, by everything that lists them
export const EVENTS_SHOWN = 100;

export const eventsOptions = () =>
  queryOptions({
    ...getV1GrendelEventsOptions({ query: { limit: EVENTS_SHOWN } }),
    select: newestFirst<Event>,
  });

// followEvents is how often the header fetches the events again: often while an event is running, now and then while none is
export const followEvents = (query: Query<Event[], any, any, any>) =>
  query.state.data?.some(isRunning) ? 5_000 : 30_000;

// useRefresh fetches again everything the page is showing, and nothing it is not
export function useRefresh() {
  const queryClient = useQueryClient();
  return useCallback(() => queryClient.refetchQueries({ type: "active" }), [queryClient]);
}
