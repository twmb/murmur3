// +build go1.5,amd64

package murmur3

// Sum32 returns the murmur3 sum of data. It is equivalent to the following
// sequence (without the extra burden and the extra allocation):
//     hasher := New32()
//     hasher.Write(data)
//     return hasher.Sum32()
func Sum32(data []byte) (h1 uint32)

// SeedSum32 returns the murmur3 sum of data with the digest initialized to
// seed.
func SeedSum32(seed uint32, data []byte) (h1 uint32)

// StringSum32 is the string version of Sum32.
func StringSum32(data string) (h1 uint32)

// SeedStringSum32 is the string version of SeedSum32.
func SeedStringSum32(seed uint32, data string) (h1 uint32)
