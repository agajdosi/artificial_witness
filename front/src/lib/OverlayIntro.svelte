<script lang="ts">
  import Typed from "typed.js";
  import { onMount, onDestroy } from "svelte";
  import { t, locale } from 'svelte-i18n';
  import { createEventDispatcher } from 'svelte';

  const dispatch = createEventDispatcher();
  function closeIntro() {
    dispatch('toggleIntro', { introVisible: false });
  }

  let typedInstance: Typed | null = null;

  let tutorialSteps: string[] = [];
  let mounted = false;

  $: tutorialSteps = [
    $t('overlayIntro.1'),
    $t('overlayIntro.2'),
    $t('overlayIntro.3'),
    $t('overlayIntro.4'),
    $t('overlayIntro.5'),
    $t('overlayIntro.6')
  ];

  function initTyped(strings: string[]) {
    const el = typeof document !== 'undefined' ? document.getElementById('typed') : null;
    if (!el) return;
    if (typedInstance) {
      typedInstance.destroy();
      typedInstance = null;
    }
    typedInstance = new Typed(el as Element, {
      strings,
      typeSpeed: 60,
      backSpeed: 10,
      fadeOut: true,
      loop: false,
      cursorChar: "|",
    });
  }

  onMount(() => {
    mounted = true;
    // ensure DOM rendered
    setTimeout(() => initTyped(tutorialSteps), 0);
  });

  // Re-create typed instance when locale (and therefore translations) changes,
  // but only after component is mounted and the element exists.
  $: if (mounted) {
    // guard so we don't call before DOM ready
    setTimeout(() => initTyped(tutorialSteps), 0);
  }

  onDestroy(() => {
    if (typedInstance) {
      typedInstance.destroy();
      typedInstance = null;
    }
  });


</script>

<div class="overlay">
  <p id="content">
    <span id="typed"></span>
  </p>
  <button on:click={closeIntro}>{$t('overlayIntro.play')}</button>
</div>

<style>
  .overlay {
    position: absolute;
    z-index: 100;
    top: 0;
    left: 0;
    height: 100vh;
    width: 100vw;
    background-color: rgba(27, 38, 54, 1);
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    color: white;
    text-align: center;
    font-size: 1.5rem;
  }

  #content {
    min-height: 20rem;
    margin: 2rem;
  }

</style>
