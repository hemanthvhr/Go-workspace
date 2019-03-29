import csv, os, math

#where the data is located
data_folder = '/home/hemanth/Downloads/taxi_release_data/taxi_log_2008_by_id/'
files_to_read = 20
ouput_data_file = 'subset_data.csv'

radiusE = 6371
error_threshold = 0.1

output_write = open(ouput_data_file, 'w')
writer = csv.writer(output_write)

rangeKm, plat, plong = 0.0, 0.0, 0.0 

plat, plong = input('Enter the coordinates of center point for your map - ')
rangeKm = input('Input the range from the center point - ')
#files_to_read = input('Enter from how many files to extract the field - ')


def cos12(a1, a2):
	return (math.cos(a1))*(math.cos(a2))

def sin12(a1, a2):
	return (math.sin(a1))*(math.sin(a2))

def distance(lat1, long1, lat2, long2):
	return radiusE*(math.sqrt(2 - 2*cos12(lat1, lat2)*(math.cos(long1-long2))-2*sin12(lat1, lat2)))

files = os.listdir(data_folder)

i = 0

for filename in files:
	if i >= files_to_read:
		break
	file_data = open(data_folder+filename)
	file_reader = csv.reader(file_data)
	for row in file_reader:
		lat = float(row[3])
		longi = float(row[2])
#		print('dist- ', distance(lat, longi, plat, plong))
		if abs(distance(lat, longi, plat, plong) - rangeKm) < error_threshold:
			writer.writerow(row) 
	file_data.close()
	i += 1

output_write.close()