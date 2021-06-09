/*
Copyright The Helm Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package registry // import "helm.sh/helm/v3/internal/experimental/registry"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/containerd/containerd/content"
	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	orascontent "github.com/oras-project/oras-go/pkg/content"
	"github.com/pkg/errors"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

const (
	// CacheRootDir is the root directory for a cache
	CacheRootDir = "cache"
)

type (
	// Cache handles local/in-memory storage of Helm charts, compliant with OCI Layout
	Cache struct {
		debug       bool
		out         io.Writer
		rootDir     string
		ociStore    *orascontent.OCIStore
		memoryStore *orascontent.Memorystore
	}

	// CacheRefSummary contains as much info as available describing a chart reference in cache
	// Note: fields here are sorted by the order in which they are set in FetchReference method
	CacheRefSummary struct {
		Name         string
		Repo         string
		Tag          string
		Exists       bool
		Manifest     *ocispec.Descriptor
		Config       *ocispec.Descriptor
		ContentLayer *ocispec.Descriptor
		Size         int64
		Digest       digest.Digest
		CreatedAt    time.Time
		Chart        *chart.Chart
	}
)

// NewCache returns a new OCI Layout-compliant cache with config
func NewCache(opts ...CacheOption) (*Cache, error) {
	cache := &Cache{
		out: ioutil.Discard,
	}
	for _, opt := range opts {
		opt(cache)
	}
	// validate
	if cache.rootDir == "" {
		return nil, errors.New("must set cache root dir on initialization")
	}
	return cache, nil
}

// FetchReference retrieves a chart ref from cache
func (cache *Cache) FetchReference(ref *Reference) (*CacheRefSummary, error) {
	if err := cache.init(); err != nil {
		return nil, err
	}
	r := CacheRefSummary{
		Name: ref.FullName(),
		Repo: ref.Repo,
		Tag:  ref.Tag,
	}
	for _, desc := range cache.ociStore.ListReferences() {
		if desc.Annotations[ocispec.AnnotationRefName] == r.Name {
			r.Exists = true
			manifestBytes, err := cache.fetchBlob(&desc)
			if err != nil {
				return &r, err
			}
			var manifest ocispec.Manifest
			err = json.Unmarshal(manifestBytes, &manifest)
			if err != nil {
				return &r, err
			}
			r.Manifest = &desc
			r.Config = &manifest.Config
			numLayers := len(manifest.Layers)
			if numLayers != 1 {
				return &r, errors.New(
					fmt.Sprintf("manifest does not contain exactly 1 layer (total: %d)", numLayers))
			}
			var contentLayer *ocispec.Descriptor
			for _, layer := range manifest.Layers {
				switch layer.MediaType {
				case HelmChartContentLayerMediaType:
					contentLayer = &layer
				}
			}
			if contentLayer == nil {
				return &r, errors.New(
					fmt.Sprintf("manifest does not contain a layer with mediatype %s", HelmChartContentLayerMediaType))
			}
			if contentLayer.Size == 0 {
				return &r, errors.New(
					fmt.Sprintf("manifest layer with mediatype %s is of size 0", HelmChartContentLayerMediaType))
			}
			r.ContentLayer = contentLayer
			info, err := cache.ociStore.Info(ctx(cache.out, cache.debug), contentLayer.Digest)
			if err != nil {
				return &r, err
			}
			r.Size = info.Size
			r.Digest = info.Digest
			r.CreatedAt = info.CreatedAt
			contentBytes, err := cache.fetchBlob(contentLayer)
			if err != nil {
				return &r, err
			}
			ch, err := loader.LoadArchive(bytes.NewBuffer(contentBytes))
			if err != nil {
				return &r, err
			}
			r.Chart = ch
		}
	}
	return &r, nil
}

// DeleteReference deletes a chart ref from cache
// TODO: garbage collection, only manifest removed
func (cache *Cache) DeleteReference(ref *Reference) (*CacheRefSummary, error) {
	if err := cache.init(); err != nil {
		return nil, err
	}
	r, err := cache.FetchReference(ref)
	if err != nil || !r.Exists {
		return r, err
	}
	cache.ociStore.DeleteReference(r.Name)
	err = cache.ociStore.SaveIndex()
	return r, err
}

// AddManifest provides a manifest to the cache index.json
func (cache *Cache) AddManifest(ref *Reference, manifest *ocispec.Descriptor) error {
	if err := cache.init(); err != nil {
		return err
	}
	cache.ociStore.AddReference(ref.FullName(), *manifest)
	err := cache.ociStore.SaveIndex()
	return err
}

// Provider provides a valid containerd Provider
func (cache *Cache) Provider() content.Provider {
	return content.Provider(cache.ociStore)
}

// Ingester provides a valid containerd Ingester
func (cache *Cache) Ingester() content.Ingester {
	return content.Ingester(cache.ociStore)
}

// ProvideIngester provides a valid oras ProvideIngester
func (cache *Cache) ProvideIngester() orascontent.ProvideIngester {
	return orascontent.ProvideIngester(cache.ociStore)
}

// init creates files needed necessary for OCI layout store
func (cache *Cache) init() error {
	if cache.ociStore == nil {
		ociStore, err := orascontent.NewOCIStore(cache.rootDir)
		if err != nil {
			return err
		}
		cache.ociStore = ociStore
		cache.memoryStore = orascontent.NewMemoryStore()
	}
	return nil
}

// fetchBlob retrieves a blob from filesystem
func (cache *Cache) fetchBlob(desc *ocispec.Descriptor) ([]byte, error) {
	reader, err := cache.ociStore.ReaderAt(ctx(cache.out, cache.debug), *desc)
	if err != nil {
		return nil, err
	}
	bytes := make([]byte, desc.Size)
	_, err = reader.ReadAt(bytes, 0)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
