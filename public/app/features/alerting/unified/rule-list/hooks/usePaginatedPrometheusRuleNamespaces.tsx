import { useCallback, useMemo, useState } from 'react';

import { PromRuleGroupDTO } from 'app/types/unified-alerting-dto';

import { isLoading, useAsync } from '../../hooks/useAsync';

export function usePaginatedPrometheusGroups<TGroup extends PromRuleGroupDTO>(
  groupsGenerator: AsyncGenerator<TGroup, void, unknown>,
  pageSize: number
) {
  const [currentPage, setCurrentPage] = useState(1);
  const [groups, setGroups] = useState<TGroup[]>([]);
  const [lastPage, setLastPage] = useState<number | undefined>(undefined);

  const [{ execute: fetchMoreGroups }, groupsRequestState] = useAsync(async (groupsCount: number) => {
    let done = false;
    const currentGroups: TGroup[] = [];

    while (currentGroups.length < groupsCount) {
      const generatorResult = await groupsGenerator.next();
      if (generatorResult.done) {
        done = true;
        break;
      }
      const group = generatorResult.value;
      currentGroups.push(group);
    }

    if (done) {
      const groupsTotal = groups.length + currentGroups.length;
      setLastPage(Math.ceil(groupsTotal / pageSize));
    }

    setGroups((groups) => [...groups, ...currentGroups]);
  });

  // lastPage could be computed from groups.length and pageSize
  const fetchInProgress = isLoading(groupsRequestState);
  const canMoveForward = !fetchInProgress && (!lastPage || currentPage < lastPage);
  const canMoveBackward = currentPage > 1 && !fetchInProgress;

  const nextPage = useCallback(async () => {
    if (canMoveForward) {
      setCurrentPage((page) => page + 1);
    }
  }, [canMoveForward]);

  const previousPage = useCallback(async () => {
    if (canMoveBackward) {
      setCurrentPage((page) => page - 1);
    }
  }, [canMoveBackward]);

  // groups.length - pageSize to have one more page loaded to prevent flickering with loading state
  // lastPage === undefined because 0 is falsy but a value which should stop fetching (e.g for broken data sources)
  const shouldFetchNextPage = groups.length - pageSize < pageSize * currentPage && lastPage === undefined;

  if (shouldFetchNextPage && !fetchInProgress) {
    fetchMoreGroups(pageSize);
  }
  const pageNamespaces = useMemo(() => {
    const pageGroups = groups.slice((currentPage - 1) * pageSize, currentPage * pageSize);
    // groupRulesByFileName mutates the array and RTKQ query freezes the response data
    return structuredClone(pageGroups);
  }, [groups, currentPage, pageSize]);

  return { isLoading: fetchInProgress, page: pageNamespaces, nextPage, previousPage, canMoveForward, canMoveBackward };
}
