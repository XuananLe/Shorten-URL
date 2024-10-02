<script lang="ts">
    import { onMount } from "svelte";
    import { Input } from "$lib/components/ui/input";
    import { Card } from "$lib/components/ui/card";
    import {
        Table,
        TableBody,
        TableCell,
        TableHead,
        TableHeader,
        TableRow,
    } from "$lib/components/ui/table";
    import { Toaster, toast } from 'svelte-sonner'
    import { Loader2, Clipboard, Trash2, ExternalLink } from "lucide-svelte";
    import { Button } from "$lib/components/ui/button";
    
    interface UrlEntry {
        id: string;
        originalUrl: string;
        shortUrl: string;
        createdAt: string;
        clickCount: number;
    }

    let url = "";
    let isLoading = false;
    let urlHistory: UrlEntry[] = [];

    onMount(() => {
        const savedHistory = localStorage.getItem('urlHistory');
        if (savedHistory) {
            urlHistory = JSON.parse(savedHistory);
        }
        console.log(urlHistory.at(0)?.clickCount)
    });

    function saveHistory() {
        localStorage.setItem('urlHistory', JSON.stringify(urlHistory));
    }

    async function shortenUrl(): Promise<void> {
        if (!url) {
            toast.error("Please enter a URL");
            return;
        }

        if (!url.startsWith('http://') && !url.startsWith('https://')) {
            url = 'https://' + url;
        }

        isLoading = true;

        try {
            await new Promise((resolve) => setTimeout(resolve, 1000));
            
            // MOCK API
            const newEntry: UrlEntry = {
                id: Math.random().toString(36).substr(2, 9),
                originalUrl: url,
                shortUrl: `https://tiny.url/${Math.random().toString(36).substr(2, 6)}`,
                createdAt: new Date().toISOString(),
                clickCount: 0
            };
            
            urlHistory = [newEntry, ...urlHistory];
            saveHistory();
            url = "";
            toast.success("URL shortened successfully!");
        } catch (e) {
            toast.error("An error occurred while shortening the URL");
        } finally {
            isLoading = false;
        }
    }

    async function copyToClipboard(shortUrl: string): Promise<void> {
        await navigator.clipboard.writeText(shortUrl);
        toast.success("Copied to clipboard!");
    }

    function deleteUrl(id: string) {
        urlHistory = urlHistory.filter(entry => entry.id !== id);
        saveHistory();
        toast.success("URL deleted from history");
    }

    function simulateClick(id: string) {
        urlHistory = urlHistory.map(entry => {
            if (entry.id === id) {
                return { ...entry, clickCount: entry.clickCount + 1 };
            }
            return entry;
        });
        saveHistory();
    }

    function formatDate(dateString: string): string {
        return new Date(dateString).toLocaleString();
    }
</script>

<Toaster />

<div class="container mx-auto p-4 max-w-4xl">
    <Card class="mb-8">
        <div class="p-6">
            <h1 class="text-3xl font-bold mb-6 text-center">URL Shortener</h1>

            <div class="flex space-x-2">
                <Input
                    type="url"
                    placeholder="Enter your long URL"
                    bind:value={url}
                />
                <Button 
                    on:click={shortenUrl} 
                    disabled={isLoading}
                    class="whitespace-nowrap"
                >
                    {#if isLoading}
                        <Loader2 class="mr-2 h-4 w-4 animate-spin" />
                    {/if}
                    Shorten URL
                </Button>
            </div>
        </div>
    </Card>

    {#if urlHistory.length > 0}
        <Card>
            <div class="p-6">
                <h2 class="text-2xl font-semibold mb-4">Your URL History</h2>
                <div class="rounded-md border">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Original URL</TableHead>
                                <TableHead>Short URL</TableHead>
                                <TableHead class="text-center">Clicks</TableHead>
                                <TableHead>Created At</TableHead>
                                <TableHead class="text-right">Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {#each urlHistory as entry}
                                <TableRow>
                                    <TableCell class="font-medium truncate max-w-[200px]" title={entry.originalUrl}>
                                        {entry.originalUrl}
                                    </TableCell>
                                    <TableCell>{entry.shortUrl}</TableCell>
                                    <TableCell class="text-center">{entry.clickCount}</TableCell>
                                    <TableCell>{formatDate(entry.createdAt)}</TableCell>
                                    <TableCell class="text-right">
                                        <div class="flex justify-end space-x-2">
                                            <Button
                                                variant="outline"
                                                size="icon"
                                                on:click={() => copyToClipboard(entry.shortUrl)}
                                            >
                                                <Clipboard class="h-4 w-4" />
                                            </Button>
                                            <Button
                                                variant="outline"
                                                size="icon"
                                                on:click={() => {
                                                    simulateClick(entry.id);
                                                    window.open(entry.originalUrl, '_blank');
                                                }}
                                            >
                                                <ExternalLink class="h-4 w-4" />
                                            </Button>
                                            <Button
                                                variant="outline"
                                                size="icon"
                                                on:click={() => deleteUrl(entry.id)}
                                            >
                                                <Trash2 class="h-4 w-4" />
                                            </Button>
                                        </div>
                                    </TableCell>
                                </TableRow>
                            {/each}
                        </TableBody>
                    </Table>
                </div>
            </div>
        </Card>
    {/if}
</div>

