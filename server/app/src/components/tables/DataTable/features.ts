import { RankingInfo, rankItem } from "@tanstack/match-sorter-utils";
import {
  columnFacetingFeature,
  columnFilteringFeature,
  columnResizingFeature,
  columnSizingFeature,
  columnVisibilityFeature,
  createFacetedRowModel,
  createFacetedUniqueValues,
  createFilteredRowModel,
  createSortedRowModel,
  FilterFn,
  filterFn_arrHas,
  filterFn_includesString,
  globalFilteringFeature,
  rowSortingFeature,
  sortFn_alphanumeric,
  sortFn_basic,
  sortFn_datetime,
  sortFn_text,
  tableFeatures
} from "@tanstack/react-table";

// Ranks the cell against the query and records the score, so a filtered row
// can report how well it matched.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const fuzzyFilter: FilterFn<any, any> = (row, columnId, value, addMeta) => {
  const itemRank = rankItem(row.getValue(columnId), value);
  addMeta?.({ itemRank });
  return itemRank.passed;
};

// Table v9 stitches features in statically rather than shipping one table that
// can do everything, so this is the single place that decides what the tables
// in this directory are capable of — and all that ends up in the bundle.
//
// The row models and the function registries live here too. A registered name
// becomes a valid value for `filterFn` / `globalFilterFn` / `sortFn` on a
// column, and an unregistered one is a type error rather than a silent
// fallback at runtime.
export const features = tableFeatures({
  columnFacetingFeature,
  columnFilteringFeature,
  columnSizingFeature,
  columnResizingFeature,
  columnVisibilityFeature,
  globalFilteringFeature,
  rowSortingFeature,
  filteredRowModel: createFilteredRowModel(),
  facetedRowModel: createFacetedRowModel(),
  facetedUniqueValues: createFacetedUniqueValues(),
  sortedRowModel: createSortedRowModel(),
  filterFns: {
    fuzzy: fuzzyFilter,
    // Facet columns hold a single string and the filter value is the list of
    // selected values, so the match is equality against that list. v8's
    // arrIncludesSome happened to work on strings via String.includes (which
    // also made it a substring match); v9 rejects non-array cell values
    // outright, which filtered every row away.
    arrHas: filterFn_arrHas,
    // Used when a column is filtered without naming a filter function.
    includesString: filterFn_includesString
  },
  // Columns do not name a sort function, so sorting picks one by inspecting
  // the data. Every name it can land on has to be registered here.
  sortFns: {
    alphanumeric: sortFn_alphanumeric,
    text: sortFn_text,
    datetime: sortFn_datetime,
    basic: sortFn_basic
  },
  // Replaces the module augmentation of FilterMeta that v8 required.
  filterMeta: {} as { itemRank: RankingInfo }
});

export type DataTableFeatures = typeof features;
