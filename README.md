# Izpi

A path tracer implemented in Golang built on top of [Peter Shirley's Raytracing books](https://raytracing.github.io).

Currently supports:

* Rendering into a float64 image buffer.
* Direct, indirect and Image-based lighting.
* Primitives: Spheres, boxes, rectangles and triangles.
* Materials: Glass, metal, Lambert, Perlin noise.
* Textures: PNG (LDR) and Radiance (HDR).

![A studio with two spheres: one glass, one metal](./images/studio.png "Sky dome test")

![The Stanford dragon in a Cornell box](./images/dragon.png "Stanford dragon")
