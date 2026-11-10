import { useAutoAnimate } from "@formkit/auto-animate/react";
import {
  Badge,
  Button,
  Center,
  Loader,
  Menu,
  Stack,
  Text,
  Tooltip
} from "@mantine/core";
import { Eye, EyeSlash, MagnifyingGlass, Trash } from "@phosphor-icons/react";
import {
  HistoryItem,
  isPending,
  removeResults,
  setActiveItem
} from "../../state/historySlice";
import { AppDispatch, useAppDispatch, useAppSelector } from "../../state/store";
import { conditionalAttribute } from "../../utils/attribute-helper";
import classes from "./SidebarButton.module.css";

export default function History() {
  const history = useAppSelector((store) => store.history.items);
  const activeTS = useAppSelector((store) => store.history.active);
  const dispatch = useAppDispatch();
  const [parent] = useAutoAnimate(/* optional config */);

  return (
    <Stack gap="xs" ref={parent}>
      {history.length > 0 ? (
        history.map((item: HistoryItem) => (
          <HistoryCard
            activeTS={activeTS}
            key={item.timestamp.toString()}
            item={item}
            dispatch={dispatch}
          />
        ))
      ) : (
        <Center>
          <Text c="dimmed" size="sm">
            History is a mystery.
          </Text>
        </Center>
      )}
    </Stack>
  );
}

type Props = {
  activeTS: number | undefined;
  item: HistoryItem;
  dispatch: AppDispatch;
};

function HistoryCard({ activeTS, item, dispatch }: Props) {
  const isActive = activeTS === item.timestamp;

  const status = () => {
    if (isPending(item)) {
      return <Loader color="blue" size="xs" />;
    }

    // No results arrived and none can now; the query is still there to run
    // again, and the menu can delete it.
    if (item.failed) {
      return (
        <Badge color="red" radius="sm" size="sm" variant="light">
          FAILED
        </Badge>
      );
    }

    return (
      <Badge color="blue" radius="sm" size="sm" variant="light">
        {`${item.results?.length ?? 0} RESULTS`}
      </Badge>
    );
  };

  return (
    <Menu shadow="md">
      <Menu.Target>
        <Tooltip label={item.query} openDelay={1_000}>
          <Button
            classNames={{
              root: classes.root,
              inner: classes.inner,
              section: classes.section,
              label: classes.label
            }}
            {...conditionalAttribute("active", isActive)}
            radius="sm"
            variant="outline"
            fullWidth
            leftSection={<MagnifyingGlass size={18} weight="bold" />}
            rightSection={status()}>
            {item.query}
          </Button>
        </Tooltip>
      </Menu.Target>

      <Menu.Dropdown>
        {!isActive ? (
          <Menu.Item
            leftSection={<Eye size={18} weight="bold" />}
            onClick={() => dispatch(setActiveItem(item))}>
            Show Results
          </Menu.Item>
        ) : (
          <Menu.Item
            leftSection={<EyeSlash size={18} weight="bold" />}
            onClick={() => dispatch(setActiveItem(null))}>
            Hide Results
          </Menu.Item>
        )}

        <Menu.Item
          color="red"
          leftSection={<Trash size={18} weight="bold" />}
          onClick={() => dispatch(removeResults(item.timestamp))}>
          Delete item
        </Menu.Item>
      </Menu.Dropdown>
    </Menu>
  );
}
