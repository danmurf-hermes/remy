<script>
  import { onMount } from 'svelte'
  import { entities, relationships } from '../lib/stores.js'
  import { getEntities, getRelationships } from '../lib/wails.js'

  let loading = true
  let selectedEntity = null
  let error = null

  onMount(async () => {
    await loadData()
  })

  async function loadData() {
    loading = true
    try {
      const [entityData, relData] = await Promise.all([getEntities(), getRelationships()])
      entities.set(entityData)
      relationships.set(relData)
    } catch (err) {
      error = err.message || String(err)
    } finally {
      loading = false
    }
  }

  function getEntityColor(type) {
    const colors = {
      person: '#007aff',
      ai: '#34c759',
      language: '#ff9500',
      technology: '#af52de',
      concept: '#ff3b30',
      place: '#5ac8fa',
      organization: '#ff2d55',
    }
    return colors[type] || '#86868b'
  }

  function selectEntity(entity) {
    selectedEntity = selectedEntity && selectedEntity.id === entity.id ? null : entity
  }

  function getRelatedEntities(entityId) {
    if (!$relationships) { return [] }
    return $relationships.filter(
      (r) => r.source_entity === entityId || r.target_entity === entityId
    )
  }

  // Simple layout: arrange entities in a circle
  $: layout = computeLayout($entities)

  function computeLayout(entityList) {
    if (!entityList || entityList.length === 0) { return [] }
    const cx = 250
    const cy = 200
    const radius = 140
    return entityList.map((e, i) => {
      const angle = (2 * Math.PI * i) / entityList.length - Math.PI / 2
      return {
        ...e,
        x: cx + radius * Math.cos(angle),
        y: cy + radius * Math.sin(angle),
      }
    })
  }

  $: edgePaths = computeEdgePaths($relationships, layout)

  function computeEdgePaths(relList, nodeList) {
    if (!relList || !nodeList) { return [] }
    const nodeMap = {}
    for (const n of nodeList) {
      nodeMap[n.name] = n
    }
    return relList
      .map((r) => {
        const src = nodeMap[r.source_entity]
        const tgt = nodeMap[r.target_entity]
        if (!src || !tgt) { return null }
        return { ...r, x1: src.x, y1: src.y, x2: tgt.x, y2: tgt.y }
      })
      .filter(Boolean)
  }
</script>

<div class="entity-graph">
  {#if error}
    <div class="error-banner">{error}</div>
  {/if}

  {#if loading}
    <div class="loading">Loading entities...</div>
  {:else if $entities.length === 0}
    <div class="empty-state">
      <p>No entities yet. Entities are extracted during deep consolidation.</p>
    </div>
  {:else}
    <div class="graph-container">
      <svg viewBox="0 0 500 400" class="graph-svg">
        <!-- Edges -->
        {#each edgePaths as edge}
          <line
            x1={edge.x1}
            y1={edge.y1}
            x2={edge.x2}
            y2={edge.y2}
            class="edge-line"
            stroke-width="1.5"
          />
          <text
            x={(edge.x1 + edge.x2) / 2}
            y={(edge.y1 + edge.y2) / 2 - 6}
            class="edge-label"
          >
            {edge.relationship}
          </text>
        {/each}

        <!-- Nodes -->
        {#each layout as node}
          <g
            class="entity-node"
            class:selected={selectedEntity && selectedEntity.id === node.id}
            role="button"
            tabindex="0"
            on:click={() => selectEntity(node)}
            on:keydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); selectEntity(node); } }}
          >
            <rect
              x={node.x - 50}
              y={node.y - 16}
              width="100"
              height="32"
              rx="16"
              ry="16"
              fill={getEntityColor(node.type)}
              opacity="0.9"
            />
            <text
              x={node.x}
              y={node.y + 5}
              class="node-label"
              text-anchor="middle"
              fill="white"
            >
              {node.name}
            </text>
          </g>
        {/each}
      </svg>
    </div>

    {#if selectedEntity}
      <div class="details-panel">
        <h3>{selectedEntity.name}</h3>
        <div class="detail-row">
          <span class="detail-label">Type:</span>
          <span class="type-badge" style="background: {getEntityColor(selectedEntity.type)};">
            {selectedEntity.type}
          </span>
        </div>
        {#if selectedEntity.description}
          <div class="detail-row">
            <span class="detail-label">Description:</span>
            <span class="detail-value">{selectedEntity.description}</span>
          </div>
        {/if}
        {#if getRelatedEntities(selectedEntity.name).length > 0}
          <div class="detail-row">
            <span class="detail-label">Relationships:</span>
            <ul class="rel-list">
              {#each getRelatedEntities(selectedEntity.name) as rel}
                <li>
                  {rel.source_entity} → {rel.target_entity}
                  <span class="rel-type">({rel.relationship})</span>
                </li>
              {/each}
            </ul>
          </div>
        {/if}
      </div>
    {/if}
  {/if}
</div>

<style>
  .entity-graph {
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .error-banner {
    padding: 8px 12px;
    background: #fff2f0;
    border: 1px solid #ffccc7;
    border-radius: 6px;
    color: #cf1322;
    font-size: 13px;
  }

  .loading {
    text-align: center;
    color: #86868b;
    padding: 40px 0;
    font-size: 14px;
  }

  .empty-state {
    text-align: center;
    color: #86868b;
    padding: 40px 20px;
    font-size: 14px;
  }

  .graph-container {
    background: #fafafa;
    border: 1px solid #e0e0e0;
    border-radius: 10px;
    overflow: hidden;
  }

  .graph-svg {
    width: 100%;
    height: auto;
  }

  .edge-line {
    stroke: #c0c0c5;
    stroke-dasharray: 4;
  }

  .edge-label {
    font-size: 10px;
    fill: #86868b;
    text-anchor: middle;
    pointer-events: none;
  }

  .entity-node {
    cursor: pointer;
    transition: opacity 0.15s;
  }

  .entity-node:hover {
    opacity: 1;
  }

  .entity-node:not(.selected) {
    opacity: 0.85;
  }

  .entity-node.selected rect {
    stroke: #1d1d1f;
    stroke-width: 2;
  }

  .node-label {
    font-size: 12px;
    font-weight: 600;
    pointer-events: none;
  }

  .details-panel {
    background: white;
    border: 1px solid #e0e0e0;
    border-radius: 10px;
    padding: 16px;
  }

  .details-panel h3 {
    margin: 0 0 12px 0;
    font-size: 16px;
    font-weight: 600;
  }

  .detail-row {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    margin-bottom: 8px;
    font-size: 13px;
  }

  .detail-label {
    color: #86868b;
    font-weight: 500;
    flex-shrink: 0;
    min-width: 80px;
  }

  .detail-value {
    color: #1d1d1f;
  }

  .type-badge {
    font-size: 11px;
    color: white;
    padding: 2px 10px;
    border-radius: 10px;
    font-weight: 500;
  }

  .rel-list {
    margin: 0;
    padding: 0 0 0 16px;
    list-style: disc;
  }

  .rel-list li {
    margin-bottom: 4px;
    font-size: 13px;
    color: #1d1d1f;
  }

  .rel-type {
    color: #86868b;
    font-size: 12px;
  }
</style>
