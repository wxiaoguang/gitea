import {queryElemChildren} from '../utils/dom.js';

export function initRepositorySearch() {
  const repositorySearchForm = document.querySelector('#repo-search-form');
  if (!repositorySearchForm) return;

  repositorySearchForm.addEventListener('change', (e) => {
    if (e.target.matches('input[type="radio"]')) {
      repositorySearchForm.submit();
    }
  });
  const filterDropdown = repositorySearchForm.querySelector('.ui.dropdown.repo-state-filter');
  filterDropdown.querySelector('.menu > .item.clear-repo-state-filter').addEventListener('click', () => {
    queryElemChildren(filterDropdown, '.menu > .item input[type="radio"]', (el) => el.checked = false);
    repositorySearchForm.submit();
  });
}
