export class NodeSetError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "NodeSetError";
  }
}

export class ParseNodeSetError extends NodeSetError {
  constructor(message: string) {
    super(message);
    this.name = "ParseNodeSetError";
  }
}

export class InvalidNodeSetError extends NodeSetError {
  constructor(message: string) {
    super(message);
    this.name = "InvalidNodeSetError";
  }
}

export class RangeSetError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "RangeSetError";
  }
}

export class ParseRangeSetError extends RangeSetError {
  constructor(message: string) {
    super(message);
    this.name = "ParseRangeSetError";
  }
}

export class InvalidRangeSetError extends RangeSetError {
  constructor(message: string) {
    super(message);
    this.name = "InvalidRangeSetError";
  }
}

export class MismatchedDimensionsError extends RangeSetError {
  constructor(message: string) {
    super(message);
    this.name = "MismatchedDimensionsError";
  }
}

export class NotImplementedError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "NotImplementedError";
  }
}
