import {
  Badge,
  Box,
  Button,
  Center,
  CloseButton,
  Divider,
  Group,
  Popover,
  Text,
  TextInput,
  useMantineColorScheme
} from "@mantine/core";
import { Column, RowData } from "@tanstack/react-table";
import { DataTableFeatures } from "./features";
import { MagnifyingGlass, PlusCircle } from "@phosphor-icons/react";
import { useState } from "react";
import classes from "./ToolbarFacetFilter.module.css";

const stringContains = (first: string, second: string): boolean => {
  return first.toLowerCase().includes(second.toLowerCase());
};

interface FacetFilterProps<TData extends RowData> {
  placeholder: string;
  column: Column<DataTableFeatures, TData, unknown>;
  Entry: React.FC<FacetEntryProps>;
}

export interface FacetEntryProps {
  entry: string;
  selected: boolean;
  onClick: (entry: string) => void;
}

export default function ToolbarFacetFilter<TData extends RowData>({
  placeholder,
  column,
  Entry
}: FacetFilterProps<TData>) {
  const [filter, setFilter] = useState("");
  const [opened, setOpened] = useState(false);

  const options = Array.from(column.getFacetedUniqueValues().keys());
  const filteredOptions = options.filter((x) => stringContains(x, filter));

  const filterValue = (column.getFilterValue() ?? []) as string[];

  const { colorScheme } = useMantineColorScheme();

  const buttonColor =
    colorScheme === "dark"
      ? filterValue.length > 0
        ? "blue.2"
        : "dark.0"
      : filterValue.length > 0
        ? "blue.4"
        : "gray.7";

  return (
    <Popover
      width={200}
      trapFocus
      position="bottom-start"
      withArrow
      shadow="md"
      opened={opened}
      onChange={setOpened}
      styles={{ dropdown: { padding: 0 } }}>
      <Popover.Target>
        <Button
          variant="default"
          className={classes.button}
          size="xs"
          leftSection={<PlusCircle size={16} />}
          color={buttonColor}
          onClick={() => setOpened((o) => !o)}>
          {placeholder}
          {filterValue.length > 0 && (
            <>
              <Divider my={4} ml={8} mr={4} orientation="vertical" />
              <Badge hiddenFrom="md" color={buttonColor} radius="sm" size="xs">
                {filterValue.length}
              </Badge>
              <Box visibleFrom="md">
                {filterValue.length > 2 ? (
                  <Badge color={buttonColor} radius="sm" size="xs">
                    {filterValue.length} selected
                  </Badge>
                ) : (
                  filterValue.map((value, index) => (
                    <Badge
                      key={index}
                      radius="sm"
                      color={buttonColor}
                      size="xs"
                      ml={4}>
                      {value}
                    </Badge>
                  ))
                )}
              </Box>
            </>
          )}
        </Button>
      </Popover.Target>
      <Popover.Dropdown>
        <Group justify="space-between" className={classes.header}>
          <Text fw="normal" size="xs" c="dark">
            Filter {placeholder}
          </Text>
          <CloseButton
            onClick={() => setOpened(false)}
            aria-label="Close modal"
            color="dark"
            iconSize={12}
          />
        </Group>
        <TextInput
          data-autofocus
          leftSection={<MagnifyingGlass weight="bold" />}
          className={classes.search}
          variant="unstyled"
          value={filter}
          size="xs"
          onChange={(e) => setFilter(e.currentTarget.value)}
          placeholder="Filter..."
          rightSection={
            column.getIsFiltered() && (
              <Button
                style={{ marginRight: 25, fontWeight: "normal" }}
                size="compact-xs"
                variant="default"
                color="blue"
                disabled={filterValue.length === 0}
                onClick={() => {
                  column.setFilterValue([]);
                  setFilter("");
                }}>
                Reset
              </Button>
            )
          }
        />

        <div className={classes.container}>
          {/* These lists hold at most a few dozen facet values, and the
              popover's scroll element does not exist until it opens, which
              @tanstack/react-virtual cannot measure. Rendering them directly
              costs nothing and keeps the list visible. */}
          {filteredOptions.map((option) => {
            const selected = filterValue.includes(option);
            return (
              <Entry
                key={option}
                entry={option}
                selected={selected}
                onClick={(entry) =>
                  selected
                    ? column.setFilterValue(
                        filterValue.filter((x) => x !== option)
                      )
                    : column.setFilterValue([...filterValue, entry])
                }
              />
            );
          })}

          {filteredOptions.length === 0 && (
            <Center py="xs">
              <Text c="dimmed" size="xs">
                No Results
              </Text>
            </Center>
          )}
        </div>
      </Popover.Dropdown>
    </Popover>
  );
}
