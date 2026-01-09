<svelte:head>
    <title>Artificial Witness | How AI Interprets People</title>
</svelte:head>

<script lang="ts">
    import { goto } from '$app/navigation';
    import { t } from 'svelte-i18n';
    import { currentGame } from '$lib/stores';
    import MenuTop from '$lib/MenuTop.svelte';
    import Navigation from '$lib/Navigation.svelte';

    function newGame() {        
        goto('/new_game');
    }
    function continueGame() {
        goto('/play');
    }
</script>

<header>
    <MenuTop/>
</header>

<h1>Artificial Witness</h1>

<div class="menu">
    <p>{$t('home.intro')}</p>
    <button on:click={newGame}>{$t('buttons.newGame')}</button>
    <button disabled={$currentGame.uuid == ""} on:click={continueGame}>{$t('buttons.continueGame')}</button>
</div>

<footer>
    <Navigation/>
    <div>
        <a href="https://github.com/agajdosi/artificial_witness">2024-2026 <span class="horflip">©</span></a>
        <a href="https://gajdosik.org">Andy Gajdosik</a>
    </div>
</footer>
