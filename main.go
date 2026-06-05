package main

import (
	"fmt"
	sync "sync"
)

// Resource represents a cached resource (e.g., texture, mesh, shader)
type Resource struct {
	Name     string
	RefCount int
}

// ResourceCache manages the loaded resources and handles eviction
type ResourceCache struct {
	mu        sync.Mutex
	resources map[string]*Resource
}

var cache = &ResourceCache{
	resources: make(map[string]*Resource),
}

// GetResource gets or loads a resource, incrementing its reference count
func (c *ResourceCache) GetResource(name string) *Resource {
	c.mu.Lock()
	defer c.mu.Unlock()
	if res, exists := c.resources[name]; exists {
		res.RefCount++
		return res
	}
	res := &Resource{Name: name, RefCount: 1}
	c.resources[name] = res
	return res
}

// ReleaseResource decrements the reference count and evicts the resource if it reaches zero
func (c *ResourceCache) ReleaseResource(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if res, exists := c.resources[name]; exists {
		res.RefCount--
		if res.RefCount <= 0 {
			delete(c.resources, name)
			fmt.Printf("Evicted resource from cache: %s\n", name)
		}
	}
}

// Node represents a node in the scene tree hierarchy
type Node struct {
	Name      string
	Resources []string
	Children  []*Node
}

// Free recursively releases the node and its children, releasing their resources
func (n *Node) Free() {
	for _, child := range n.Children {
		child.Free()
	}
	for _, resName := range n.Resources {
		cache.ReleaseResource(resName)
	}
	fmt.Printf("Freed node: %s\n", n.Name)
}

// SceneTree manages the active scene and transitions
type SceneTree struct {
	CurrentScene *Node
}

// ChangeScene transitions to a new scene, properly freeing the old scene hierarchy
func (st *SceneTree) ChangeScene(newScene *Node) {
	if st.CurrentScene != nil {
		fmt.Printf("Transitioning scene. Freeing old scene: %s\n", st.CurrentScene.Name)
		st.CurrentScene.Free()
	}
	st.CurrentScene = newScene
	fmt.Printf("Loaded new scene: %s\n", newScene.Name)
}

func main() {
	fmt.Println("Starting SceneTree Resource Management Simulation...")

	tree := &SceneTree{}

	// Create Scene A (lightweight main menu)
	sceneA := &Node{
		Name:      "SceneA (Main Menu)",
		Resources: []string{"menu_theme.ogg"},
	}
	cache.GetResource("menu_theme.ogg")
	tree.ChangeScene(sceneA)

	// Create Scene B (heavy gameplay scene)
	sceneB := &Node{
		Name:      "SceneB (Level 1)",
		Resources: []string{"level1_map.mesh", "heavy_texture.png"},
		Children: []*Node{
			{
				Name:      "Enemy",
				Resources: []string{"enemy_model.mesh", "heavy_texture.png"},
			},
		},
	}
	cache.GetResource("level1_map.mesh")
	cache.GetResource("heavy_texture.png")
	cache.GetResource("enemy_model.mesh")
	cache.GetResource("heavy_texture.png")

	// Transition to Scene B
	tree.ChangeScene(sceneB)

	// Transition back to Scene A (should free Scene B and its unique resources)
	sceneA2 := &Node{
		Name:      "SceneA (Main Menu)",
		Resources: []string{"menu_theme.ogg"},
	}
	cache.GetResource("menu_theme.ogg")
	tree.ChangeScene(sceneA2)

	// Verify cache state
	cache.mu.Lock()
	fmt.Printf("Active resources in cache: %d\n", len(cache.resources))
	for name, res := range cache.resources {
		fmt.Printf(" - %s (RefCount: %d)\n", name, res.RefCount)
	}
	cache.mu.Unlock()
}
