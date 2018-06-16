import matplotlib.pyplot as plt
import numpy as np
import cv2
from mpl_toolkits.mplot3d import Axes3D
from sklearn.cluster import KMeans

nbcolor = 3

data = np.genfromtxt('./data.csv', delimiter=',', skip_header=1, dtype=int,
                     skip_footer=0, names=['r', 'g', 'b'])

ndata = np.array([list(tup) for tup in data])

kmeans = KMeans(n_clusters=nbcolor)
kmeans.fit(ndata)

centers = kmeans.cluster_centers_
labels = kmeans.labels_

colors = centers.astype(int)

for c in colors:
    r = c[0]
    g = c[1]
    b = c[2]
    str = "{0} {1} {2}".format(r,g,b)
    print(str)

