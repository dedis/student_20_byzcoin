#! /bin/python
import matplotlib.pyplot as plt
import pandas as pd
import os
import datetime

data_dir = './test_data/'
files = [(data_dir + fname) for fname in os.listdir(data_dir)\
         if fname.startswith('coins') and fname.endswith('.csv')]

def read_all_files(files):
    df = pd.DataFrame()
    for fname in files:
        data = pd.read_csv(fname)
        # We need to add variables regarding
        # batching and keeping here.
        batch = '_batch' in fname
        keep = not '_nokeep' in fname
        rowcount = len(data.index)
        b_vals = pd.Series([batch for i in range(rowcount)])
        k_vals = pd.Series([keep for i in range(rowcount)])
        data = data.assign(batch=b_vals.values)
        data = data.assign(keep=k_vals.values)
        # If the data frame is empty (first iteration),
        # we append no matter what. Otherwise, we append
        # IFF the colums are the same.
        if df.empty \
           or (len(data.columns) == len(df.columns) \
           and (data.columns == df.columns).all()):
            df = df.append(data, ignore_index=True)
    return df

df = read_all_files(files)
delays = list(set(df['delay']))
keep = list(set(df['keep']))
batch = list(set(df['batch']))

def plot_columns(ax, columns):
    for i in range(len(columns)):
        column = columns[i]
        d = data[column]
        ax.bar(i, d, label=column[:-9])

    ax.set_xlabel('transactions / batch size')
    ax.legend(bbox_to_anchor=(0.1, 0.5), loc='center left',)

# Monitoring tree
# - send
# - prepare
#   - prepare_intro
#   - create_tx
#   - sign
# - confirm

for delay in delays:
    for k in keep:
        for b in batch:
            titlestring =  'Byzcoin Simulation Monitor'
            # No whitespace, colons or commata in filenames
            namestring = (str(datetime.datetime.now()) + '-' + titlestring).replace(' ','').replace(':','-')

            data = df.loc[df['delay'] == delay].sort_values('hosts')
            data = data.loc[data['keep'] == k]
            data = data.loc[data['batch'] == b]
            data = data.reset_index()

            fig, axs = plt.subplots(2, 3)
            fig.set_size_inches(16, 10)

            fig.suptitle(titlestring)

            #plot_columns(axs[0], ['send_user_sum',
            #                      'prepare_user_sum',
            #                      'confirm_user_sum'])
            data.plot.bar(\
                    x='hosts',\
                    y= ['send_wall_sum',
                                  'prepare_wall_sum',
                                  'confirm_wall_sum'],\
                    stacked=True, ax=axs[0][0])

            axs[0][0].set_ylabel('Time in seconds')
            
            data.plot.bar(x='hosts', y=['prepare.prepare_intro_wall_sum',
                                  'prepare.create_tx_wall_sum',
                                  'prepare.sign_wall_sum'], stacked=True, ax=axs[0][1])

            data.plot.bar(x='hosts', y=['create_state_change_wall_sum'], stacked=True, ax=axs[0][2])

            data.plot.bar(x='hosts', y=['process_one_tx_wall_sum'], stacked=True, ax=axs[1][0])

            data.plot.bar(x='hosts', y=['p_o_t.init_wall_sum', 'p_o_t.execute_wall_sum', 'p_o_t.increment_wall_sum', 'p_o_t.verify_wall_sum', 'p_o_t.store_wall_sum'], stacked=True, ax=axs[1][2])

            [[ax.set_ylim([0, 12]) for ax in a] for a in axs]

            plt.savefig(data_dir + namestring + '.png')
            plt.close()
