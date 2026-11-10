import { ScrollArea, Table, Text } from "@mantine/core";
import { useDebouncedValue } from "@mantine/hooks";
import {
  ColumnDef,
  flexRender,
  Row,
  RowData,
  Table as TableType,
  useTable
} from "@tanstack/react-table";

import { useVirtualizer } from "@tanstack/react-virtual";
import { ReactNode, useRef, useState } from "react";
import { BookDetail } from "../../../state/messages";
import { conditionalAttribute } from "../../../utils/attribute-helper";
import classes from "./DataTable.module.css";
import { DataTableFeatures, features } from "./features";
import Toolbar from "./Toolbar";

interface DataTableProps<TData extends RowData> {
  // ColumnDef is invariant in its value type and every column carries its
  // own, so this is the array type that columnHelper.columns() produces.
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  columns: ColumnDef<DataTableFeatures, TData, any>[];
  data: TData[];
  facetFilters?: (table: TableType<DataTableFeatures, TData>) => ReactNode[];
}

export default function DataTable<TData extends RowData>({
  columns,
  data,
  facetFilters
}: DataTableProps<TData>) {
  // Table virtualization stuff
  const virtualizedRef = useRef<HTMLDivElement>(null);

  const [searchQuery, setSearchQuery] = useState<string>("");
  const [debouncedSearchQuery] = useDebouncedValue<string>(searchQuery, 250);

  const table = useTable({
    features,
    data: data,
    columns: columns,
    enableMultiSort: true,
    state: {
      globalFilter: debouncedSearchQuery
    },
    defaultColumn: {
      minSize: 50,
      size: Number.MAX_SAFE_INTEGER,
      maxSize: Number.MAX_SAFE_INTEGER
    },
    onGlobalFilterChange: (value) => setSearchQuery((value as string) ?? ""),
    globalFilterFn: "fuzzy",
    enableFilters: true,
    enableGlobalFilter: true,
    columnResizeMode: "onChange"
  });

  const { rows: tableRows } = table.getRowModel();

  const rowVirtualizer = useVirtualizer({
    count: tableRows.length,
    getScrollElement: () => virtualizedRef.current,
    estimateSize: () => 50,
    overscan: 10
  });

  const virtualItems = rowVirtualizer.getVirtualItems();

  const paddingTop =
    virtualItems.length > 0 ? virtualItems?.[0]?.start || 0 : 0;
  const paddingBottom =
    virtualItems.length > 0
      ? rowVirtualizer.getTotalSize() -
        (virtualItems?.[virtualItems.length - 1]?.end || 0)
      : 0;
  return (
    <>
      <Toolbar
        table={table}
        searchQuery={searchQuery}
        setSearchQuery={setSearchQuery}
        facetFilters={facetFilters}
      />
      <ScrollArea
        viewportRef={virtualizedRef}
        className={classes.container}
        type="hover"
        scrollbarSize={6}
        classNames={{
          thumb: classes.thumb
        }}
        offsetScrollbars={false}>
        <Table highlightOnHover verticalSpacing="0" fz="xs">
          <Table.Thead className={classes.head}>
            {table.getHeaderGroups().map((headerGroup) => (
              <Table.Tr key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <Table.Th
                    key={header.id}
                    pl="lg"
                    py="sm"
                    className={classes.headerCell}
                    style={{
                      width:
                        header.getSize() === Number.MAX_SAFE_INTEGER
                          ? "auto"
                          : header.getSize()
                    }}>
                    {flexRender(
                      header.column.columnDef.header,
                      header.getContext()
                    )}
                    <div
                      role="presentation"
                      onMouseDown={header.getResizeHandler()}
                      onTouchStart={header.getResizeHandler()}
                      {...conditionalAttribute(
                        "resizing",
                        header.column.getIsResizing()
                      )}
                      className={classes.resizer}></div>
                  </Table.Th>
                ))}
              </Table.Tr>
            ))}
          </Table.Thead>
          <Table.Tbody>
            {paddingTop > 0 && (
              <Table.Tr>
                <Table.Td style={{ height: `${paddingTop}px` }} />
              </Table.Tr>
            )}
            {rowVirtualizer.getVirtualItems().map((virtualRow) => {
              const row = tableRows[virtualRow.index] as unknown as Row<
                DataTableFeatures,
                BookDetail
              >;
              return (
                <Table.Tr key={row.id} style={{ height: 50 }}>
                  {row.getVisibleCells().map((cell) => {
                    return (
                      <Table.Td
                        key={cell.id}
                        style={{
                          width:
                            cell.column.getSize() === Number.MAX_SAFE_INTEGER
                              ? "auto"
                              : cell.column.getSize()
                        }}>
                        <Text
                          fz="xs"
                          component="div"
                          pl="xl"
                          lineClamp={1}
                          c="dark">
                          {flexRender(
                            cell.column.columnDef.cell,
                            cell.getContext()
                          )}
                        </Text>
                      </Table.Td>
                    );
                  })}
                </Table.Tr>
              );
            })}
            {paddingBottom > 0 && (
              <Table.Tr>
                <Table.Td style={{ height: `${paddingBottom}px` }} />
              </Table.Tr>
            )}
          </Table.Tbody>
        </Table>
      </ScrollArea>
    </>
  );
}
