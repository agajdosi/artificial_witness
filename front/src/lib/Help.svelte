<script lang="ts">
    import { createEventDispatcher } from 'svelte';
    import { t } from 'svelte-i18n';

    const dispatch = createEventDispatcher();

    function closeHelp() {
        dispatch('toggleHelp', { helpVisible: false });
    }

</script>

<div class="infobox_overlay" role="dialog" aria-modal="true">
    <div class="infobox">
        <h1>{$t('help.title')}</h1>
        <p>1. {$t('help.text1')}</p>
        <p>2. {$t('help.text2')}</p>
        <p>3. {$t('help.text3')}</p>
        <p>4. {$t('help.text4')}</p>
        <div class="actions">
            <button on:click={closeHelp}>{$t('buttons.close')}</button>
        </div>
    </div>
</div>

<style>
h1 {
    margin: 20px 10px 15px 10px;
    font-size: 2.4rem;
}
p {
    margin: 15px 20px;
    text-align: left;
}
button{
    margin: 20px;
}
</style>
