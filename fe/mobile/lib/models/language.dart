// Language model
class Language {
  final String name;
  final bool isActive;
  final String imageUrl;

  const Language({
    required this.name,
    required this.isActive,
    required this.imageUrl,
  });
}

const List<Language> kLanguageTracks = [
  Language(
    name: 'Javanese',
    isActive: true,
    imageUrl:
        'https://lh3.googleusercontent.com/aida-public/AB6AXuBD7LBbwJZLiCWgCAAm88kFVCou5cNic_OTa-GVd6-BCoqBugfTfK2n6TLSvOOHe_sP3-kie4y8pQi0avlBsBZZ8K0AYkoAMTjKly5J1LvBi_rOwkWcLZHsILnKvvIkPUH0I1OfDsx3QhuP7hz7AsTXjceyNnDYR3K7hMEYeV9o5bSTTjbUOkJjULYCzcMtIMEaIFbcD0GxJb2REdWPO-5UDghqcBMHedm8ED1948-vgl9qlAxJaFnu9d4pCw9i12q7JzBHojCVRwnH',
  ),
  Language(
    name: 'Sundanese',
    isActive: false,
    imageUrl:
        'https://lh3.googleusercontent.com/aida-public/AB6AXuA2Daht31EbnxsQpbNaQjUtbs5YcnnVWvkcCa8qpPUmc5BK6xKTtM35VtJ-XM_SzZCmgpuMoPXz1Kcz2b6oFwnW2cJ0Al5eqpFoWiMsk048Qlp4cQXtKXxDKvf8kiWTa7_hPRqZMcnS7k9JWuF0e4umxxXkB9gyjkSZy9nqYLRtDVCKFn3yEKrbB5AVb6vSCI8r3SCHBxIkoXZOKsHeDGEaKsW22FAIPBiKnNdHTNLCK0HU2wp_3j9AhwM6cH8lHsWTFzVDrW4jRFun',
  ),
  Language(
    name: 'Balinese',
    isActive: false,
    imageUrl:
        'https://lh3.googleusercontent.com/aida-public/AB6AXuAryQPNrtHYE4jTD9cloJxyJDLkKurzEUSUYulBYzxHpcNZ7Xcx1mxNBf3kfz4rYtX-q5hUM4nzMLtD7mgMrRKVDilpYj3NRWdLhpPp-ddPr40VvuTgbWJDcApdiMVjWTtRzih_9yXNKyyqbVCOT41t25G6hcnzNFhMDfzLAfnGfYZTtPhqrM2REiCKUaEejxzEwONpBLGysaID8UU8vtt0iEIOwRk_xxqfvH4nAtKqYFOvaNqBQNzOidx-G2M_4IaCvI5FBh3HWt05',
  ),
];
