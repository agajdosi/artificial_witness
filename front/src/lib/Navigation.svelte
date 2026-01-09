<script lang="ts">
    import { goto } from '$app/navigation';
    import { page } from '$app/state';
    import { t } from 'svelte-i18n';
</script>

<div class="navigation">
    <button onclick={() => goto('/')} class="navigation" disabled={page.url.pathname==='/'}>{$t('buttons.back')}</button>
</div>

<style>
button {
    all: unset;
    min-width: 25px;
    text-decoration: underline;
}
button:hover{
    all: unset;
    min-width: 25px;
    cursor: pointer;
    text-decoration: underline;
}
button:disabled{
    color: #666666;
}
</style>