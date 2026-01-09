<svelte:head>
    <title>Artificial Witness | Start a New Investigation</title> 
</svelte:head>

<script lang="ts">
    import { goto } from '$app/navigation';
    import { ListAvailableModels } from '$lib/main';
    import { onMount } from 'svelte';
    import { currentGame } from '$lib/stores';
    import { t } from 'svelte-i18n';
    import { selectedModel } from '$lib/stores';
    import MenuTop from '$lib/MenuTop.svelte';
    import Navigation from '$lib/Navigation.svelte';


    let models: any[] = $state([]);
    let loading = $state(true);

    onMount(async () => {
        models = await ListAvailableModels(true, "price");
        loading = false;
    });

    async function modelSelected(event: Event) {
        const target = event.target as HTMLButtonElement;
        const name = target.value;
        console.log("Starting game with model:", name);
        selectedModel.set(name);
        currentGame.set({
            uuid: '',
            level: 0,
            Score: 0,
            Model: '',
            investigation: {
                uuid: '',
                game_uuid: '',
                suspects: [],
                rounds: [],
                CriminalUUID: '',
                InvestigationOver: false,
                Timestamp: ''
            },
            GameOver: false,
            Investigator: '',
            Timestamp: ''
        });
        await goto('/play');
    }
</script>
<header>
    <MenuTop/>
</header>

<h1>{$t('new_game.title')}</h1>

<div class="services">
    {#if loading}
        Loading services...
    {:else if models.length === 0}
        No models available
    {:else}
        {#each models as model}
                <button onclick={modelSelected} value={model.Name}>{model.Model} {model.Name}</button>
        {/each}
    {/if}
</div>

<footer>
    <Navigation/>
    <div>
        <a href="https://github.com/agajdosi/artificial_witness">2024-2026 <span class="horflip">©</span></a>
        <a href="https://gajdosik.org">Andi Gajdosik</a>
    </div>
</footer>
