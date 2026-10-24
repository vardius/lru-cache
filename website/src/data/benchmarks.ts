// Results behind the benchmark charts and tables, updated together. The suite reruns with the commands on the
// Benchmarks page; the comparison comes from a harness that runs the old and new versions side by side.

export const environment = 'Apple M4 Max, Go 1.27.0, median of 5 runs';

/** The two versions the charts compare. */
export const versions = { before: 'untagged', after: 'v1' } as const;

export type Group = 'get' | 'set' | 'fill' | 'parallel';

export interface Result {
	ns: number;
	allocs: number;
}

export interface Comparison {
	scenario: string;
	before: Result;
	after: Result;
}

/**
 * The untagged version against v1 on the same harness: a cache of 1,024 entries, string keys and byte slice values,
 * one call per op. The untagged version's budget is one byte more than 1,024 values, so it holds as many entries.
 */
export const comparison: Record<Group, Comparison[]> = {
	get: [
		{ scenario: 'Hit', before: { ns: 15.0, allocs: 0 }, after: { ns: 15.5, allocs: 0 } },
		{ scenario: 'Miss', before: { ns: 8.5, allocs: 0 }, after: { ns: 9.3, allocs: 0 } },
	],
	set: [
		{ scenario: 'Existing key', before: { ns: 17.7, allocs: 0 }, after: { ns: 15.0, allocs: 0 } },
		{ scenario: 'New key, evicting', before: { ns: 69.9, allocs: 2 }, after: { ns: 15.0, allocs: 0 } },
	],
	fill: [{ scenario: 'New and 1,024 Sets', before: { ns: 68570, allocs: 2072 }, after: { ns: 18760, allocs: 3 } }],
	// About half the gets miss and set the key. The untagged version allocates for each new key: 0.9 per op.
	parallel: [
		{
			scenario: 'Read-through, 16 goroutines',
			before: { ns: 391.4, allocs: 0.9 },
			after: { ns: 253.0, allocs: 0 },
		},
	],
};

export interface SuiteResult {
	benchmark: string;
	what: string;
	ns: number;
	bytes: number;
	allocs: number;
}

/** v1's own benchmark suite, `go test -bench .`. */
export const suite: SuiteResult[] = [
	{ benchmark: 'Get', what: 'A hit, in no particular order, on a full cache', ns: 16.8, bytes: 0, allocs: 0 },
	{ benchmark: 'GetMiss', what: 'A key that isn\'t there', ns: 9.9, bytes: 0, allocs: 0 },
	{ benchmark: 'SetExisting', what: 'A new value for a cached key', ns: 15.9, bytes: 0, allocs: 0 },
	{ benchmark: 'SetNew', what: 'A new key on a full cache: evicts the least recently used', ns: 15.3, bytes: 0, allocs: 0 },
	{ benchmark: 'Fill', what: 'New, then 1,024 Sets', ns: 18730, bytes: 82000, allocs: 3 },
	{ benchmark: 'Parallel', what: 'Read-through from 16 goroutines at once: about half the gets miss and set', ns: 255.5, bytes: 0, allocs: 0 },
];
