import { RowData, Table } from "@tanstack/react-table";
import { DataTableFeatures } from "./features";
import { Button, Menu, useMantineColorScheme } from "@mantine/core";
import { Check, Sliders } from "@phosphor-icons/react";

interface ToolbarViewOptionsProps<TData extends RowData> {
  table: Table<DataTableFeatures, TData>;
}

export function ToolbarViewOptions<TData extends RowData>({
  table
}: ToolbarViewOptionsProps<TData>) {
  const { colorScheme } = useMantineColorScheme();
  const buttonColor = colorScheme === "dark" ? "blue.2" : "dark.0";
  const hideableColumns = table
    .getAllColumns()
    .filter(
      (column) =>
        typeof column.accessorFn !== "undefined" && column.getCanHide()
    );

  if (hideableColumns.length === 0) {
    return null;
  }

  return (
    <Menu
      shadow="md"
      width={200}
      closeOnItemClick={false}
      position="bottom-end">
      <Menu.Target>
        <Button
          color={buttonColor}
          size="xs"
          variant="default"
          leftSection={<Sliders />}>
          View
        </Button>
      </Menu.Target>
      <Menu.Dropdown>
        <Menu.Label>Toggle columns</Menu.Label>
        <Menu.Divider />
        {hideableColumns.map((column) => {
          return (
            <Menu.Item
              style={{ textTransform: "capitalize" }}
              key={column.id}
              leftSection={
                column.getIsVisible() ? (
                  <Check size={14} />
                ) : (
                  <div style={{ width: 14, height: 14 }} />
                )
              }
              onClick={() => column.toggleVisibility(!column.getIsVisible())}>
              {column.id}
            </Menu.Item>
          );
        })}
      </Menu.Dropdown>
    </Menu>
  );
}
