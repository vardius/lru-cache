// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
	site: 'https://vardius.github.io',
	base: '/lru-cache',
	trailingSlash: 'always',
	// Pages import the runnable programs in ../examples, so the dev server may read outside this folder.
	vite: { server: { fs: { allow: ['..'] } } },
	integrations: [
		starlight({
			title: 'lru-cache',
			description:
				'A least recently used cache for Go: generic, safe for concurrent use, exact in its eviction order, and allocation-free after New.',
			logo: { src: './src/assets/mark.svg', alt: '' },
			favicon: '/favicon.svg',
			components: { Hero: './src/components/Hero.astro' },
			// Long lines wrap instead of scrolling under the copy button on phones.
			expressiveCode: { defaultProps: { wrap: true } },
			social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/vardius/lru-cache' }],
			customCss: ['./src/styles/custom.css'],
			sidebar: [
				{
					label: 'Start here',
					items: ['getting-started'],
				},
				{
					label: 'Guide',
					items: ['get-and-set', 'eviction', 'concurrency'],
				},
				{
					label: 'Reference',
					items: [
						'benchmarks',
						'migrating-from-the-untagged-version',
						{ label: 'API reference', link: 'https://pkg.go.dev/github.com/vardius/lru-cache' },
					],
				},
			],
		}),
	],
});
