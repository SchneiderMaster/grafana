import { ReactNode, useState } from 'react';

import { locationService } from '@grafana/runtime';
import {
  Badge,
  BadgeColor,
  Card,
  EmptySearchResult,
  EmptyState,
  FilterInput,
  Icon,
  IconName,
  LinkButton,
  Stack,
  TextLink,
} from '@grafana/ui';
import { Page } from 'app/core/components/Page/Page';

import { DeleteRepositoryButton } from './DeleteRepositoryButton';
import { SyncRepository } from './SyncRepository';
import { Repository, ResourceCount } from './api';
import { NEW_URL, PROVISIONING_URL } from './constants';
import { useRepositoryList } from './hooks';

export default function RepositoryListPage() {
  const [items, isLoading] = useRepositoryList({ watch: true });

  return (
    <Page navId="provisioning" subTitle="View and manage your configured repositories">
      <Page.Contents isLoading={isLoading}>
        <RepositoryListPageContent items={items} />
      </Page.Contents>
    </Page>
  );
}

function RepositoryListPageContent({ items }: { items?: Repository[] }) {
  const [query, setQuery] = useState('');
  if (!items?.length) {
    return (
      <EmptyState
        variant="call-to-action"
        message="You haven't created any repository configs yet"
        button={
          <LinkButton icon="plus" href={NEW_URL} size="lg">
            Create repository config
          </LinkButton>
        }
      />
    );
  }

  const filteredItems = items.filter((item) => item.metadata?.name?.includes(query));

  return (
    <Stack direction={'column'} gap={3}>
      <Stack gap={2}>
        <FilterInput placeholder="Search" value={query} onChange={setQuery} />
        <LinkButton href={NEW_URL} variant="primary" icon={'plus'}>
          Add repository config
        </LinkButton>
      </Stack>
      <Stack direction={'column'}>
        {!!filteredItems.length ? (
          filteredItems.map((item) => {
            const name = item.metadata?.name ?? '';

            let icon: IconName = 'database'; // based on type
            let meta: ReactNode[] = [
              // TODO... add counts? and sync info
            ];
            switch (item.spec?.type) {
              case 'github':
                icon = 'github';
                const spec = item.spec.github;
                const url = `https://github.com/${spec?.owner}/${spec?.repository}/`;
                let branch = url;
                if (spec?.branch) {
                  branch += `tree/` + spec?.branch;
                }
                meta.push(
                  <TextLink key={'link'} external style={{ color: 'inherit' }} href={branch}>
                    {branch}
                  </TextLink>
                );

                if (item.status?.webhook?.id) {
                  const hook = url + `/settings/hooks/${item.status?.webhook?.id}`;
                  meta.push(
                    <TextLink key={'webhook'} style={{ color: 'inherit' }} href={hook}>
                      Webhook <Icon name={'check'} />
                    </TextLink>
                  );
                }
                break;

              case 'local':
                meta.push(<span key={'path'}>{item.spec.local?.path}</span>);
                break;
            }

            return (
              <Card key={name}>
                <Card.Figure>
                  <Icon name={icon} width={40} height={40} />
                </Card.Figure>
                <Card.Heading>
                  <Stack>
                    {item.spec?.title} <StatusBadge repo={item} name={name} />
                  </Stack>
                </Card.Heading>
                <Card.Description>
                  {item.spec?.description}
                  {item.status?.stats?.length && (
                    <Stack>
                      {item.status.stats.map((v) => (
                        <LinkButton fill="outline" size="md" href={getListURL(item, v)}>
                          {v.count} {v.resource}
                        </LinkButton>
                      ))}
                    </Stack>
                  )}
                </Card.Description>
                <Card.Meta>{meta}</Card.Meta>
                <Card.Actions>
                  <LinkButton href={`${PROVISIONING_URL}/${name}`} variant="secondary">
                    Manage
                  </LinkButton>
                  <LinkButton href={`${PROVISIONING_URL}/${name}/edit`} variant="secondary">
                    Edit
                  </LinkButton>
                  <SyncRepository repository={item} />
                </Card.Actions>
                <Card.SecondaryActions>
                  <DeleteRepositoryButton name={name} />
                </Card.SecondaryActions>
              </Card>
            );
          })
        ) : (
          <EmptySearchResult>No results matching your query</EmptySearchResult>
        )}
      </Stack>
    </Stack>
  );
}

// This should return a URL in the UI that will show the selected values
function getListURL(repo: Repository, stats: ResourceCount): string {
  if (stats.resource === 'playlists') {
    return '/playlists';
  }
  if (repo.spec?.folder) {
    return `/d/${repo.spec?.folder}`;
  }
  return '/dashboards';
}

interface StatusBadgeProps {
  repo: Repository;
  name: string;
}
function StatusBadge({ repo, name }: StatusBadgeProps) {
  const state = repo.status?.sync?.state ?? '';

  let tooltip: string | undefined = undefined;
  let color: BadgeColor = 'purple';
  let text = 'Unknown';
  let icon: IconName = 'exclamation-triangle';
  switch (state) {
    case 'success':
      icon = 'check';
      text = 'In sync';
      color = 'green';
      break;
    case null:
    case undefined:
    case '':
      color = 'orange';
      text = 'Pending';
      icon = 'spinner';
      tooltip = 'Waiting for health check to run';
      break;
    case 'working':
    case 'pending':
      color = 'orange';
      text = 'Syncing';
      icon = 'spinner';
      break;
    case 'error':
      color = 'red';
      text = 'Error';
      icon = 'exclamation-triangle';
      break;
    default:
      break;
  }
  return (
    <Badge
      color={color}
      icon={icon}
      text={text}
      style={{ cursor: 'pointer' }}
      tooltip={tooltip}
      onClick={() => {
        locationService.push(`${PROVISIONING_URL}/${name}/?tab=health`);
      }}
    />
  );
}
