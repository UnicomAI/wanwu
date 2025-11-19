export function transformGraphData(backendData, options = {}) {
  if (!backendData) {
    return { nodes: [], edges: [] }
  }

  const nodes = backendData.nodes || []
  const edges = backendData.edges || []
  
  const {
    getNodeId = (node, index) => node.entity_name || `n${index}`,
    getNodeLabel = (node, index) => node.entity_name || '',
    getEdgeId = (edge, index) => `e${index}`,
    getEdgeLabel = (edge, index) => edge.description || ''
  } = options

  const transformedNodes = nodes.map((node, index) => {
    const nodeId = getNodeId(node, index)
    const nodeLabel = getNodeLabel(node, index)

    // Removeafter端Datain of  size And style.fill Property，让 graphMap  of  defaultNode Configuration生效
    const { size, style, ...nodeWithoutSizeAndStyle } = node
    const nodeStyle = style && style.fill ? { ...style, fill: undefined } : style
    const cleanedNode = nodeStyle && Object.keys(nodeStyle).length > 0 
      ? { ...nodeWithoutSizeAndStyle, style: nodeStyle }
      : nodeWithoutSizeAndStyle

    return {
      id: nodeId,
      label: nodeLabel,
      type: 'circle',
      ...cleanedNode
    }
  })

  const transformedEdges = edges.map((edge, index) => {
    const edgeId = getEdgeId(edge, index)
    const edgeLabel = getEdgeLabel(edge, index)

    return {
      id: edgeId,
      source: edge.source_entity,
      target: edge.target_entity,
      label: edgeLabel,
      ...(edge.weight && {
        style: {
          lineWidth: Math.max(1, Math.min(5, edge.weight / 2))
        }
      }),
      ...edge
    }
  })

  return {
    nodes: transformedNodes,
    edges: transformedEdges
  }
}

export default {
  transformGraphData
}
