# izhikevich_neuron_model
Playing around Izhikevich spiking models (http://www.izhikevich.org/publications/spikes.pdf)




This script provides the means to play around this spiking neuron modelisation.

[neurons] is a slice containing [nNeurons].

Neurons are created with NewNIzkvch(id, type), then activated using Activate() method.

          
      	neurons[0] = NewNIzkvch(0, 2)
	neurons[0].Activate()

In order to feed something to the neuron, use it's [in] channel:

      neurons[i].in <- -1.8 * rand.Float64()


Neuron state is updated every [latencyUpdate].
If neuron's membrane potentiel reachec +30mV threshhold, a spike is induced in the [out] channel of said neuron.


Neurons can be connected by synapses using the NewSynapse(in chan, weight float64, out chan) function.

    	synapses[0] = NewSynapse(neurons[0].out, -0.2, neurons[2].in)	
	go synapses[0].Transmit()
      
Synapse will transmit incoming spike to their connected neuron by applying a weight.

main() is set up to provide a live terminal plotting interface.
Interface is not done, quit before end of simulation :)

