# Dependency Graph

The dependency graph is a top-down representation of the relationships between files in
the pre-processor.

## FR-1: Debugability of the dependency graph

The complex nature of the dependency graph (DG) makes it difficult to reason about the relationships between files and
to identify issues when processing source file(s). To address this, the DG must have extensive logging capabilities that
provides detailed information about the relationships between files.

- **FR-1.1.0 - top-down nature is clear:** The DH works in a top-down manner, starting from the lowest level of the graph
and working its way up to the highest level. 
- 



## FR-2: Scan files included by the source file

- **FR-1.1.1 dependency graph receives working directory:** the pre-processor must receive the working directory as an
  argument, which is used to resolve relative paths of included files. The pre-processor must use this working directory
  to locate and read the included files. The graph throw a fatal error if the working directory is not provided, not found,
  is not a directory or cannot be accessed.
- **FR-1.1.2 dependency graph is acyclic:** the pre-processor must ensure that the dependency graph is acyclic, meaning
  that there are no circular dependencies between files. If a circular dependency is detected, the pre-processor must
  throw a fatal error indicating that a circular dependency has been detected and stop processing.
- **FR-1.1.3 file could not be found:** if a file included by the source file cannot be found, the
  pre-processor must throw a fatal error indicating that a dependency cannot be resolved and stop
  processing.




