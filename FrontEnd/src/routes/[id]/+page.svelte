<script lang="ts">
    import { page } from "$app/stores";
    $: id = $page.params.id;
</script>

<div>
    Redirecting to {id}
</div>
