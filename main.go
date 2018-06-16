package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
)

var client http.Client = http.Client{}

// Used to represent a pixel
type Point2D struct {
	X, Y uint
}

// Used to represent RGB
type Point3D struct {
	X, Y, Z uint
}

type Centroid struct {
	Coord          Point3D
	AssignedPoints []Point3D
}

type Cluster struct {
	Carte       *image.Image
	Nb          int
	NbIteration int
	Centroids   []*Centroid
}

func check(e error) {
	if e != nil {
		log.Println(e.Error())
		panic(e)
	}
}

/**
  * @name : Centroid.AvgSum
  * @description : Sum all the 3D Component and return the averaged sum in a 3D Point
				   Used for new cluster assignation
  * @params : void
  * @return : Point3D
**/
//
func (c Centroid) AvgSum() Point3D {
	tmp := Point3D{X: 0, Y: 0, Z: 0}
	for _, p := range c.AssignedPoints {
		tmp.X += p.X
		tmp.Y += p.Y
		tmp.Z += p.Z
	}

	m := uint(len(c.AssignedPoints))
	tmp.X /= m
	tmp.Y /= m
	tmp.Z /= m

	return tmp
}

/**
  * @name : Init
  * @description : Init all the centroid by random position
  * @params :
		- nb_centroid int
		- carte *Image
  * @return : int
**/
func (cluster *Cluster) init(nb_centroid, nb_iter int, carte *image.Image) {
	cluster.Nb = nb_centroid
	cluster.Carte = carte
	cluster.Centroids = make([]*Centroid, cluster.Nb)
	cluster.NbIteration = nb_iter

	for i := 0; i < cluster.Nb; i++ {
		cluster.Centroids[i] = &Centroid{
			Coord: Point3D{
				X: uint(rand.Intn(255)),
				Y: uint(rand.Intn(255)),
				Z: uint(rand.Intn(255)),
			},
			AssignedPoints: make([]Point3D, 0),
		}
	}
}

func (cluster Cluster) distance3D(p1 *Point3D, p2 *Point3D) float64 {
	dx := float64(p1.X - p2.X)
	dy := float64(p1.Y - p2.Y)
	dz := float64(p1.Z - p2.Z)
	x := math.Pow(dx, 2)
	y := math.Pow(dy, 2)
	z := math.Pow(dz, 2)
	return math.Sqrt(x + y + z)
}

/**
  * @name : findBestCentroid
  * @description : find the closest centroid of the 3D Point (RGB)
  * @params : *Point
  * @return : int
**/

func (cluster *Cluster) findBestCentroid(p *Point3D) int {

	shortest := cluster.distance3D(p, &cluster.Centroids[0].Coord)
	idx := 0

	for i := 1; i < cluster.Nb; i++ {
		centroid := cluster.Centroids[i]
		dist := cluster.distance3D(p, &centroid.Coord)

		if dist < shortest {
			shortest = dist
			idx = i
		}
	}

	return idx
}

func (cluster *Cluster) canStop(prevCluster []Point3D, curIter int) bool {

	l := len(prevCluster)

	if l == 0 {
		return false
	}

	for i := 0; i < l; i++ {
		dist := cluster.distance3D(&cluster.Centroids[i].Coord, &prevCluster[i])
		fmt.Println("Distance from previous centroids", dist)
	}

	if cluster.NbIteration <= curIter {
		return false
	}

	return true
}

/**
  * @name : KMeans
  * @description : Apply kmeans on a 3D Points
  * @params : void
  * @return : void
**/
func (cluster *Cluster) KMeans(nb_centroid, nb_iter int, carte *image.Image) []*Centroid {

	cluster.init(nb_centroid, nb_iter, carte)

	bounds := (*cluster.Carte).Bounds()

	prevCluster := make([]Point3D, 0)

	iter := 0
	for cluster.canStop(prevCluster, iter) {

		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, g, b, _ := (*cluster.Carte).At(x, y).RGBA()
				nx := r >> 8
				ny := g >> 8
				nz := b >> 8
				curPoint := Point3D{X: uint(nx), Y: uint(ny), Z: uint(nz)}
				idx := cluster.findBestCentroid(&curPoint)
				cluster.Centroids[idx].AssignedPoints = append(
					cluster.Centroids[idx].AssignedPoints,
					curPoint)
			}
		}

		for k := 0; k < cluster.Nb; k++ {
			if len(cluster.Centroids[k].AssignedPoints) > 0 {
				cluster.Centroids[k].Coord = cluster.Centroids[k].AvgSum() // New coord
			}
		}

		iter++
		fmt.Println("Iteration ", iter)
	}

	return cluster.Centroids
}

func DownloadImage(url string) *image.Image {
	res, err := client.Get(url)

	check(err)

	defer res.Body.Close()

	content, err := ioutil.ReadAll(res.Body)
	check(err)

	img, err := jpeg.Decode(bytes.NewReader(content))
	check(err)
	return &img
}

func imageToCsv(img *image.Image) {
	f, err := os.Create("./data.csv")
	check(err)
	defer f.Close()

	bounds := (*img).Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := (*img).At(x, y).RGBA()
			r = r >> 8
			g = g >> 8
			b = b >> 8
			red := strconv.Itoa(int(r))
			green := strconv.Itoa(int(g))
			blue := strconv.Itoa(int(b))

			_, err := f.WriteString(red + "," + green + "," + blue + "\n")
			check(err)
		}
	}
}

func main() {
	urlTest := "http://res.cloudinary.com/hpcjvlhpl/image/upload/c_crop,h_200,x_269,y_224/v1529073173/dog_in_the_woods_by_svitakovaeva-d4ecywm.jpg" //"http://res.cloudinary.com/hpcjvlhpl/image/upload/v1528039592/ufqph4cnhifv6eewu6yg.jpg"
	img := DownloadImage(urlTest)
	cluster := new(Cluster)

	res := cluster.KMeans(3, 300, img)

	fmt.Println("Result", res)
}
