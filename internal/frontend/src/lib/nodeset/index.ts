export {
  InvalidNodeSetError,
  InvalidRangeSetError,
  MismatchedDimensionsError,
  NodeSetError,
  NotImplementedError,
  ParseNodeSetError,
  ParseRangeSetError,
  RangeSetError,
} from "./errors";
export { NodeSetIterator, RangeSetNDIterator } from "./iterator";
export { NodeSet } from "./node-set";
export { RangeSet, RangeSetItem, type Slice } from "./range-set";
export { RangeSetND } from "./range-set-nd";

import { NodeSet } from "./node-set";

export function parseNodeSet(input: string): NodeSet {
  return new NodeSet(input);
}

export function expand(input: string): string[] {
  return new NodeSet(input).iterator().stringSlice();
}

export function fold(nodes: readonly string[]): string {
  if (nodes.length === 0) return "";
  return new NodeSet(nodes.join(",")).toString();
}

export function contains(input: string, node: string): boolean {
  for (const n of new NodeSet(input)) {
    if (n === node) return true;
  }
  return false;
}
